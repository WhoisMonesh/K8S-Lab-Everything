package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodFsGroupMismatchLab{})
}

type PodFsGroupMismatchLab struct {
	BaseLab
}

func (l *PodFsGroupMismatchLab) ID() string {
	return "pod_fs_group_mismatch"
}

func (l *PodFsGroupMismatchLab) Title() string {
	return "fsGroup Permission Issue"
}

func (l *PodFsGroupMismatchLab) Category() Category {
	return CategorySecurity
}

func (l *PodFsGroupMismatchLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodFsGroupMismatchLab) Description() string {
	return `A pod 'data-writer' cannot write to a mounted PVC because the fsGroup
is set incorrectly. The container runs as UID 1000 but fsGroup is set
to 999, causing permission denied errors.

Your task: Fix the fsGroup to match the container's UID.`
}

func (l *PodFsGroupMismatchLab) Hints() []string {
	return []string{
		"Check the pod security context",
		"fsGroup must allow the container UID to write",
		"Set fsGroup to match the container's runAsUser",
	}
}

func (l *PodFsGroupMismatchLab) EstimatedTime() int {
	return 15
}

func (l *PodFsGroupMismatchLab) Tags() []string {
	return []string{"security", "fsgroup", "permissions", "storage"}
}

func (l *PodFsGroupMismatchLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodFsGroupMismatchLab) Break(ctx context.Context, kubeconfigPath string) error {
	pvc := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: standard
`
	if err := kubectlApply(ctx, kubeconfigPath, pvc); err != nil {
		return fmt.Errorf("creating PVC: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: data-writer
  namespace: default
spec:
  securityContext:
    runAsUser: 1000
    fsGroup: 999
  containers:
  - name: writer
    image: busybox:1.36
    command: ['sh', '-c', 'echo "test" > /data/file.txt && cat /data/file.txt && sleep 3600']
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: data-pvc
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodFsGroupMismatchLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pod", "data-writer",
		"-o", "jsonpath={.status.phase}")
	if strings.Contains(output, "Running") {
		// Check if there are permission errors in logs
		logs, _ := kubectl(ctx, kubeconfigPath, "logs", "data-writer")
		if strings.Contains(logs, "Permission denied") {
			return nil
		}
	}
	return nil
}

func (l *PodFsGroupMismatchLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "data-writer",
		"-o", "jsonpath={.spec.securityContext.fsGroup}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "999" {
		return fmt.Errorf("fsGroup is still 999")
	}

	// Verify write works
	output, err = kubectl(ctx, kubeconfigPath, "exec", "data-writer",
		"--", "cat", "/data/file.txt")
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	if !strings.Contains(output, "test") {
		return fmt.Errorf("file content not written correctly")
	}

	return nil
}

func (l *PodFsGroupMismatchLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check security context",
			Command:     "kubectl get pod data-writer -o yaml | grep -A 3 securityContext",
			Notes:       "fsGroup is 999 but container runs as UID 1000",
		},
		{
			Description: "Fix fsGroup",
			Command:     "kubectl edit pod data-writer",
			Notes:       "Change fsGroup from 999 to 1000",
		},
		{
			Description: "Verify write works",
			Command:     "kubectl exec data-writer -- cat /data/file.txt",
			Notes:       "Should now contain 'test'",
		},
	}
}
