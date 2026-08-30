package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&InitContainerFailureLab{})
}

type InitContainerFailureLab struct {
	BaseLab
}

func (l *InitContainerFailureLab) ID() string {
	return "init_container_failure"
}

func (l *InitContainerFailureLab) Title() string {
	return "Init Container Failure"
}

func (l *InitContainerFailureLab) Category() Category {
	return CategoryAppDeployment
}

func (l *InitContainerFailureLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *InitContainerFailureLab) Description() string {
	return `A pod has an init container that is failing, preventing the main
container from starting. The init container is trying to connect to a
service that doesn't exist yet.

Your task: Fix the init container configuration so the pod can start successfully.`
}

func (l *InitContainerFailureLab) Hints() []string {
	return []string{
		"Check the pod status and init container logs",
		"Init containers run before the main container",
		"The init container might be waiting for a dependency that won't come",
	}
}

func (l *InitContainerFailureLab) EstimatedTime() int {
	return 15
}

func (l *InitContainerFailureLab) Tags() []string {
	return []string{"init-container", "pod", "troubleshooting", "workloads"}
}

func (l *InitContainerFailureLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *InitContainerFailureLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: app-with-init
  namespace: default
spec:
  initContainers:
  - name: wait-for-db
    image: busybox:1.36
    command: ['sh', '-c', 'until nc -z postgres-service 5432; do echo waiting for db; sleep 2; done']
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'while true; do echo app running; sleep 15; done']
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *InitContainerFailureLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "app-with-init",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("checking pod: %w", err)
	}

	phase := strings.TrimSpace(output)
	if phase == "Pending" || phase == "" {
		return nil
	}

	return fmt.Errorf("pod is %s (expected Pending)", phase)
}

func (l *InitContainerFailureLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "app-with-init",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("checking pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod not running (phase: %s)", strings.TrimSpace(output))
	}

	return nil
}

func (l *InitContainerFailureLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod app-with-init",
			Notes:       "Pod should be in Init:0/1 or Pending state",
		},
		{
			Description: "Check init container logs",
			Command:     "kubectl logs app-with-init -c wait-for-db",
			Notes:       "See what the init container is waiting for",
		},
		{
			Description: "Fix: Remove the init container dependency",
			Command:     `kubectl patch pod app-with-init --type='json' -p='[{"op":"remove","path":"/spec/initContainers"}]'`,
			Notes:       "Or create the postgres-service if it's needed",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod app-with-init",
			Notes:       "Pod should transition to Running state",
		},
	}
}
