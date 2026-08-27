package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ContainerLifecyclePrestartLab{})
}

type ContainerLifecyclePrestartLab struct {
	BaseLab
}

func (l *ContainerLifecyclePrestartLab) ID() string {
	return "container_lifecycle_prestart"
}

func (l *ContainerLifecyclePrestartLab) Title() string {
	return "PreStart Hook Failing"
}

func (l *ContainerLifecyclePrestartLab) Category() Category {
	return CategoryWorkloads
}

func (l *ContainerLifecyclePrestartLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ContainerLifecyclePrestartLab) Description() string {
	return `A pod 'init-app' is failing to start because its PreStart lifecycle
hook is failing. The hook runs a command that doesn't exist in the container.

Your task: Fix the PreStart hook command so the pod can start.`
}

func (l *ContainerLifecyclePrestartLab) Hints() []string {
	return []string{
		"Check pod events for lifecycle hook failures",
		"PreStart hooks run before the container's main process starts",
		"Verify the command exists in the container image",
	}
}

func (l *ContainerLifecyclePrestartLab) EstimatedTime() int {
	return 15
}

func (l *ContainerLifecyclePrestartLab) Tags() []string {
	return []string{"lifecycle", "prestart", "hooks", "workloads"}
}

func (l *ContainerLifecyclePrestartLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ContainerLifecyclePrestartLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: init-app
  namespace: default
spec:
  containers:
  - name: app
    image: nginx:alpine
    lifecycle:
      preStop:
        exec:
          command: ["/bin/sh", "-c", "echo pre-stop executed"]
    readinessProbe:
      httpGet:
        path: /
        port: 80
      initialDelaySeconds: 5
      periodSeconds: 5
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *ContainerLifecyclePrestartLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ContainerLifecyclePrestartLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "init-app",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "init-app",
		"-o", "jsonpath={.status.conditions[?(@.type=='Ready')].status}")
	if err != nil {
		return fmt.Errorf("failed to check conditions: %w", err)
	}

	if strings.TrimSpace(output) != "True" {
		return fmt.Errorf("pod is not ready")
	}

	return nil
}

func (l *ContainerLifecyclePrestartLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod init-app",
			Notes:       "Pod might be in CrashLoopBackOff or CreateError",
		},
		{
			Description: "Check pod events",
			Command:     "kubectl describe pod init-app | grep -A 10 Events",
			Notes:       "Look for lifecycle hook failures",
		},
		{
			Description: "Fix the lifecycle hook",
			Command:     "kubectl edit pod init-app",
			Notes:       "Ensure the preStop command uses a valid executable",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod init-app",
			Notes:       "Pod should be Running and Ready",
		},
	}
}
