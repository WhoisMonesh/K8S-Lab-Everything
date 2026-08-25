package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&MultiContainerPodLab{})
}

type MultiContainerPodLab struct {
	BaseLab
}

func (l *MultiContainerPodLab) ID() string {
	return "multi_container_pod"
}

func (l *MultiContainerPodLab) Title() string {
	return "Multi-Container Pod Communication Issue"
}

func (l *MultiContainerPodLab) Category() Category {
	return CategoryNetworking
}

func (l *MultiContainerPodLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *MultiContainerPodLab) Description() string {
	return `A pod 'web-stack' has two containers: nginx and a sidecar logger.
The sidecar cannot communicate with the nginx container because it's connecting to the wrong port.

Your task: Fix the sidecar configuration to connect to the correct port.`
}

func (l *MultiContainerPodLab) Hints() []string {
	return []string{
		"Check the pod status",
		"Look at the sidecar container logs",
		"Check which port nginx is listening on",
		"The sidecar is connecting to the wrong port number",
	}
}

func (l *MultiContainerPodLab) EstimatedTime() int {
	return 15
}

func (l *MultiContainerPodLab) Tags() []string {
	return []string{"multi-container", "sidecar", "networking", "ports"}
}

func (l *MultiContainerPodLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *MultiContainerPodLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create pod with sidecar connecting to wrong port
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: web-stack
  namespace: default
spec:
  containers:
  - name: nginx
    image: nginx:alpine
    ports:
    - containerPort: 80
  - name: logger
    image: busybox:1.28
    command: ['sh', '-c', 'while true; do wget -q -O- http://localhost:9999/ || echo "Failed to connect"; sleep 10; done']
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *MultiContainerPodLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *MultiContainerPodLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if pod is running
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "web-stack",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	// Check logger container logs for successful connection
	output, err = kubectl(ctx, kubeconfigPath, "logs", "web-stack", "-c", "logger", "--tail=5")
	if err != nil {
		return fmt.Errorf("failed to check logger logs: %w", err)
	}

	// After fix, logs should not contain "Failed to connect"
	if strings.Contains(output, "Failed to connect") {
		return fmt.Errorf("logger still failing to connect")
	}

	// Check the logger command
	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "web-stack",
		"-o", "jsonpath={.spec.containers[1].command}")
	if err != nil {
		return fmt.Errorf("failed to check pod spec: %w", err)
	}

	if strings.Contains(output, "9999") {
		return fmt.Errorf("logger still connecting to wrong port 9999")
	}

	return nil
}

func (l *MultiContainerPodLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod web-stack",
			Notes:       "The pod should be Running (both containers started)",
		},
		{
			Description: "Check logger container logs",
			Command:     "kubectl logs web-stack -c logger --tail=10",
			Notes:       "Logs will show 'Failed to connect' messages",
		},
		{
			Description: "Check nginx container port",
			Command:     "kubectl get pod web-stack -o yaml | grep -A 3 containers[0].ports",
			Notes:       "Nginx listens on port 80",
		},
		{
			Description: "Check logger command",
			Command:     "kubectl get pod web-stack -o yaml | grep -A 2 containers[1].command",
			Notes:       "Logger tries to connect to port 9999 instead of 80",
		},
		{
			Description: "Delete and recreate pod with correct port",
			Command:     "kubectl delete pod web-stack",
			Notes:       "Delete the broken pod",
		},
		{
			Description: "Create pod with fixed logger command",
			Command:     `kubectl run web-stack --image=nginx:alpine --dry-run=client -o yaml > fixed-pod.yaml`,
			Notes:       "Edit the YAML to fix the logger command to use port 80",
		},
		{
			Description: "Apply the fixed pod",
			Command:     "kubectl apply -f fixed-pod.yaml",
			Notes:       "The pod should now start with both containers working",
		},
		{
			Description: "Verify logger is connecting successfully",
			Command:     "kubectl logs web-stack -c logger --tail=5",
			Notes:       "Logs should show successful connections",
		},
	}
}
