package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodEmptyDirSizeLimitLab{})
}

type PodEmptyDirSizeLimitLab struct {
	BaseLab
}

func (l *PodEmptyDirSizeLimitLab) ID() string {
	return "pod_empty_dir_size_limit"
}

func (l *PodEmptyDirSizeLimitLab) Title() string {
	return "EmptyDir Size Limit Exceeded"
}

func (l *PodEmptyDirSizeLimitLab) Category() Category {
	return CategoryStorage
}

func (l *PodEmptyDirSizeLimitLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodEmptyDirSizeLimitLab) Description() string {
	return `A pod 'cache-app' uses an EmptyDir volume with a 100Mi size limit.
The application writes more data than this, causing the pod to be
evicted due to disk pressure.

Your task: Increase the EmptyDir size limit to prevent eviction.`
}

func (l *PodEmptyDirSizeLimitLab) Hints() []string {
	return []string{
		"Check the EmptyDir volume configuration",
		"sizeLimit controls the maximum size of the EmptyDir",
		"Increase sizeLimit to accommodate application data",
	}
}

func (l *PodEmptyDirSizeLimitLab) EstimatedTime() int {
	return 10
}

func (l *PodEmptyDirSizeLimitLab) Tags() []string {
	return []string{"emptydir", "size-limit", "eviction", "storage"}
}

func (l *PodEmptyDirSizeLimitLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodEmptyDirSizeLimitLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: cache-app
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'for i in $(seq 1 200); do dd if=/dev/zero of=/cache/file$i bs=1M count=1; done && sleep 3600']
    volumeMounts:
    - name: cache
      mountPath: /cache
  volumes:
  - name: cache
    emptyDir:
      sizeLimit: 100Mi
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodEmptyDirSizeLimitLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(30 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pod", "cache-app",
		"-o", "jsonpath={.status.reason}")
	if strings.Contains(output, "Evicted") {
		return nil
	}
	return nil
}

func (l *PodEmptyDirSizeLimitLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "cache-app",
		"-o", "jsonpath={.spec.volumes[0].emptyDir.sizeLimit}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "100Mi" {
		return fmt.Errorf("sizeLimit is still 100Mi")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "cache-app",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod status: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *PodEmptyDirSizeLimitLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check EmptyDir configuration",
			Command:     "kubectl get pod cache-app -o yaml | grep -A 3 emptyDir",
			Notes:       "sizeLimit is 100Mi which is too small",
		},
		{
			Description: "Fix sizeLimit",
			Command:     "kubectl edit pod cache-app",
			Notes:       "Change sizeLimit from 100Mi to 500Mi or higher",
		},
		{
			Description: "Verify pod runs without eviction",
			Command:     "kubectl get pod cache-app",
			Notes:       "Pod should stay in Running state",
		},
	}
}
