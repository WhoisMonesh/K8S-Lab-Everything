package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&EphemeralContainer{})
}

type EphemeralContainer struct {
	BaseLab
}

func (l *EphemeralContainer) ID() string            { return "ephemeral_container" }
func (l *EphemeralContainer) Title() string         { return "Debug Pod with Ephemeral Container" }
func (l *EphemeralContainer) Category() Category    { return CategoryWorkloads }
func (l *EphemeralContainer) Difficulty() Difficulty { return DifficultyMedium }
func (l *EphemeralContainer) EstimatedTime() int    { return 20 }
func (l *EphemeralContainer) Tags() []string        { return []string{"ephemeral", "debugging", "workloads"} }

func (l *EphemeralContainer) Description() string {
	return `A pod is running but the application inside is misbehaving.
Add an ephemeral debug container to investigate the issue without restarting the pod.`
}

func (l *EphemeralContainer) Hints() []string {
	return []string{
		"Use kubectl debug to add an ephemeral container",
		"Check the ephemeral container status",
		"Execute commands inside the debug container",
	}
}

func (l *EphemeralContainer) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *EphemeralContainer) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: debug-target
spec:
  replicas: 1
  selector:
    matchLabels:
      app: debug-target
  template:
    metadata:
      labels:
        app: debug-target
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *EphemeralContainer) Verify(ctx context.Context, kubeconfigPath string) error {
	pods, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=debug-target",
		"-o", "jsonpath={.items[0].metadata.name}")
	if err != nil {
		return err
	}
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", pods,
		"-o", "jsonpath={.status.ephemeralContainerStatuses[*].state}")
	if err != nil {
		return err
	}
	if output == "" {
		return fmt.Errorf("no ephemeral container found")
	}
	return nil
}

func (l *EphemeralContainer) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Get pod name", Command: "kubectl get pods -l app=debug-target"},
		{Description: "Add ephemeral container", Command: "kubectl debug -it <pod-name> --image=busybox:1.36 --target=app"},
		{Description: "Debug the application", Command: "Use nslookup, wget, or other tools to investigate"},
	}
}
