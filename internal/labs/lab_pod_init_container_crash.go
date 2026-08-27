package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodInitContainerCrashLab{})
}

type PodInitContainerCrashLab struct {
	BaseLab
}

func (l *PodInitContainerCrashLab) ID() string {
	return "pod_init_container_crash"
}

func (l *PodInitContainerCrashLab) Title() string {
	return "Init Container Crash Loop"
}

func (l *PodInitContainerCrashLab) Category() Category {
	return CategoryWorkloads
}

func (l *PodInitContainerCrashLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodInitContainerCrashLab) Description() string {
	return `A pod 'data-loader' is stuck in Init:CraLoopBackOff. The init container
is supposed to download configuration files but crashes immediately.

The init container runs a wget command that fails due to an incorrect URL.

Your task: Fix the init container so the pod can start successfully.`
}

func (l *PodInitContainerCrashLab) Hints() []string {
	return []string{
		"Check the init container logs with kubectl logs <pod> -c <init-container>",
		"Look at the wget URL in the init container command",
		"The URL scheme might be wrong (http vs https)",
	}
}

func (l *PodInitContainerCrashLab) EstimatedTime() int {
	return 15
}

func (l *PodInitContainerCrashLab) Tags() []string {
	return []string{"init-containers", "crashloop", "troubleshooting"}
}

func (l *PodInitContainerCrashLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodInitContainerCrashLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: data-loader
  namespace: default
spec:
  initContainers:
  - name: fetch-config
    image: busybox:1.36
    command: ['sh', '-c', 'wget -q -O /config/app.yaml htp://nonexistent.internal/config.yaml || exit 1']
    volumeMounts:
    - name: config-vol
      mountPath: /config
  containers:
  - name: app
    image: nginx:alpine
    volumeMounts:
    - name: config-vol
      mountPath: /config
  volumes:
  - name: config-vol
    emptyDir: {}
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodInitContainerCrashLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pod", "data-loader",
		"-o", "jsonpath={.status.initContainerStatuses[0].state}")
	if !strings.Contains(output, "waiting") && !strings.Contains(output, "CrashLoop") {
		return fmt.Errorf("expected init container to be crashing")
	}
	return nil
}

func (l *PodInitContainerCrashLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "data-loader",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *PodInitContainerCrashLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod data-loader",
			Notes:       "Pod should show Init:CraLoopBackOff",
		},
		{
			Description: "Check init container logs",
			Command:     "kubectl logs data-loader -c fetch-config",
			Notes:       "wget fails due to invalid URL scheme 'htp://'",
		},
		{
			Description: "Fix the init container URL",
			Command:     "kubectl edit pod data-loader",
			Notes:       "Change 'htp://' to 'http://' or 'https://' in the wget command",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod data-loader",
			Notes:       "Pod should now be in Running state",
		},
	}
}
