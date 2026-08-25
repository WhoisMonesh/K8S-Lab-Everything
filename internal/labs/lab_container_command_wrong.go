package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ContainerCommandWrongLab{})
}

type ContainerCommandWrongLab struct {
	BaseLab
}

func (l *ContainerCommandWrongLab) ID() string {
	return "container_command_wrong"
}

func (l *ContainerCommandWrongLab) Title() string {
	return "Container Command Causing CrashLoop"
}

func (l *ContainerCommandWrongLab) Category() Category {
	return CategoryWorkloads
}

func (l *ContainerCommandWrongLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *ContainerCommandWrongLab) Description() string {
	return `A pod 'monitoring' is in CrashLoopBackOff because the container command is wrong.
The container is trying to execute a non-existent binary.

Your task: Fix the container command to use the correct executable.`
}

func (l *ContainerCommandWrongLab) Hints() []string {
	return []string{
		"Check the pod logs",
		"Look at the container's command/args",
		"The command references a binary that doesn't exist",
		"Fix the command to use a valid executable",
	}
}

func (l *ContainerCommandWrongLab) EstimatedTime() int {
	return 10
}

func (l *ContainerCommandWrongLab) Tags() []string {
	return []string{"command", "args", "crashloop", "troubleshooting"}
}

func (l *ContainerCommandWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ContainerCommandWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create pod with wrong command
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: monitoring
  namespace: default
spec:
  containers:
  - name: agent
    image: busybox:1.28
    command: ["/usr/local/bin/nonexistent-binary"]
    args: ["--config", "/etc/config.yaml"]
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *ContainerCommandWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ContainerCommandWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if pod is running
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "monitoring",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	// Check restart count
	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "monitoring",
		"-o", "jsonpath={.status.containerStatuses[0].restartCount}")
	if err != nil {
		return fmt.Errorf("failed to check restart count: %w", err)
	}

	restarts := strings.TrimSpace(output)
	if restarts != "0" {
		return fmt.Errorf("pod has restarts (count: %s)", restarts)
	}

	return nil
}

func (l *ContainerCommandWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod monitoring",
			Notes:       "The pod should be in CrashLoopBackOff",
		},
		{
			Description: "Check pod logs",
			Command:     "kubectl logs monitoring --previous",
			Notes:       "Look for 'exec format error' or 'command not found' error",
		},
		{
			Description: "Check the pod spec",
			Command:     "kubectl get pod monitoring -o yaml | grep -A 3 command",
			Notes:       "The command references /usr/local/bin/nonexistent-binary",
		},
		{
			Description: "Delete and recreate pod with correct command",
			Command:     "kubectl delete pod monitoring",
			Notes:       "Delete the broken pod",
		},
		{
			Description: "Create pod with correct command",
			Command:     `kubectl run monitoring --image=busybox:1.28 -- sleep 3600`,
			Notes:       "Use a simple command that keeps the pod running",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod monitoring",
			Notes:       "The pod should now be in Running state",
		},
	}
}
