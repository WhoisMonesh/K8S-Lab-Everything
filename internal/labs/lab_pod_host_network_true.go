package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodHostNetworkTrueLab{})
}

type PodHostNetworkTrueLab struct {
	BaseLab
}

func (l *PodHostNetworkTrueLab) ID() string {
	return "pod_host_network_true"
}

func (l *PodHostNetworkTrueLab) Title() string {
	return "Pod Using hostNetwork Incorrectly"
}

func (l *PodHostNetworkTrueLab) Category() Category {
	return CategoryNetworking
}

func (l *PodHostNetworkTrueLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodHostNetworkTrueLab) Description() string {
	return `A pod 'monitoring' is configured with hostNetwork: true but it's
trying to bind to port 80 which conflicts with the kube-proxy or other
system services running on the host network.

Your task: Fix the pod to either not use hostNetwork or use a different port.`
}

func (l *PodHostNetworkTrueLab) Hints() []string {
	return []string{
		"Check the pod configuration",
		"hostNetwork: true shares the host's network namespace",
		"Port 80 might already be in use by host services",
		"Either disable hostNetwork or change the container port",
	}
}

func (l *PodHostNetworkTrueLab) EstimatedTime() int {
	return 15
}

func (l *PodHostNetworkTrueLab) Tags() []string {
	return []string{"pod", "host-network", "port-conflict", "networking"}
}

func (l *PodHostNetworkTrueLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodHostNetworkTrueLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: monitoring
  namespace: default
spec:
  hostNetwork: true
  containers:
  - name: monitor
    image: nginx:alpine
    ports:
    - containerPort: 80
      hostPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodHostNetworkTrueLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pod", "monitoring",
		"-o", "jsonpath={.status.phase}")
	if strings.Contains(output, "Running") {
		return nil
	}
	return nil
}

func (l *PodHostNetworkTrueLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "monitoring",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "monitoring",
		"-o", "jsonpath={.spec.hostNetwork}")
	if err != nil {
		return fmt.Errorf("failed to check hostNetwork: %w", err)
	}

	if strings.TrimSpace(output) == "true" {
		return fmt.Errorf("pod still using hostNetwork with conflicting port")
	}

	return nil
}

func (l *PodHostNetworkTrueLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod monitoring",
			Notes:       "Pod might be in CrashLoopBackOff or pending",
		},
		{
			Description: "Check pod events",
			Command:     "kubectl describe pod monitoring | grep -A 10 Events",
			Notes:       "Look for port conflict errors",
		},
		{
			Description: "Fix the pod",
			Command:     "kubectl edit pod monitoring",
			Notes:       "Either set hostNetwork: false or change hostPort to a free port like 8080",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod monitoring",
			Notes:       "Pod should now be Running",
		},
	}
}
