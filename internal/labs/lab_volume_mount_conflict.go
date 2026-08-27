package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&VolumeMountConflict{})
}

type VolumeMountConflict struct {
	BaseLab
}

func (l *VolumeMountConflict) ID() string             { return "volume_mount_conflict" }
func (l *VolumeMountConflict) Title() string          { return "Volume Mount Path Conflict" }
func (l *VolumeMountConflict) Category() Category     { return CategoryStorage }
func (l *VolumeMountConflict) Difficulty() Difficulty { return DifficultyMedium }
func (l *VolumeMountConflict) EstimatedTime() int     { return 15 }
func (l *VolumeMountConflict) Tags() []string         { return []string{"storage", "volumes", "mounts"} }

func (l *VolumeMountConflict) Description() string {
	return `A pod is failing because two containers are trying to mount the same path.
Fix the volume mount configuration to resolve the conflict.`
}

func (l *VolumeMountConflict) Hints() []string {
	return []string{
		"Check volume mount paths",
		"Look for conflicting mountPath values",
		"Change one container's mount path",
	}
}

func (l *VolumeMountConflict) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *VolumeMountConflict) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Pod
metadata:
  name: mount-conflict
spec:
  containers:
  - name: app1
    image: nginx:alpine
    volumeMounts:
    - name: shared
      mountPath: /data
  - name: app2
    image: nginx:alpine
    volumeMounts:
    - name: shared
      mountPath: /data
  volumes:
  - name: shared
    emptyDir: {}`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *VolumeMountConflict) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "mount-conflict",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return err
	}
	if output == "Error" || output == "CrashLoopBackOff" {
		return fmt.Errorf("pod in error state")
	}
	return nil
}

func (l *VolumeMountConflict) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod spec", Command: "kubectl get pod mount-conflict -o yaml"},
		{Description: "Fix mount paths", Command: "kubectl edit pod mount-conflict"},
		{Description: "Change app2 mountPath", Command: "Change app2 mountPath from /data to /data2"},
	}
}
