package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&CNIFailureRecoveryLab{})
}

type CNIFailureRecoveryLab struct {
	BaseLab
}

func (l *CNIFailureRecoveryLab) ID() string {
	return "cni_failure_recovery"
}

func (l *CNIFailureRecoveryLab) Title() string {
	return "CNI Plugin Failure Recovery"
}

func (l *CNIFailureRecoveryLab) Category() Category {
	return CategoryTroubleshooting
}

func (l *CNIFailureRecoveryLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *CNIFailureRecoveryLab) Description() string {
	return `The CNI plugin configuration has been corrupted on a node, causing all
new pods on that node to fail networking. Pods remain in Pending state
with network-related errors.

Your task: Diagnose the CNI configuration issue and restore network
connectivity for pods on the affected node.`
}

func (l *CNIFailureRecoveryLab) Hints() []string {
	return []string{
		"Check if CNI pods are running in kube-system namespace",
		"Look at the CNI configuration files on the affected node",
		"Check /etc/cni/net.d/ on the node for configuration files",
	}
}

func (l *CNIFailureRecoveryLab) EstimatedTime() int {
	return 30
}

func (l *CNIFailureRecoveryLab) Tags() []string {
	return []string{"cni", "networking", "troubleshooting", "node", "cluster"}
}

func (l *CNIFailureRecoveryLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CNIFailureRecoveryLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: cni-test
  namespace: default
spec:
  containers:
  - name: test
    image: busybox:1.36
    command: ['sh', '-c', 'while true; do echo ok; sleep 5; done']
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return err
	}

	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("getting control plane node: %w", err)
	}

	if _, err := dockerExec(ctx, nodeName, "bash", "-c", "rm -rf /etc/cni/net.d/*"); err != nil {
		return fmt.Errorf("removing CNI config: %w", err)
	}

	return nil
}

func (l *CNIFailureRecoveryLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "cni-test",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return nil
	}

	phase := strings.TrimSpace(output)
	if phase == "Pending" || phase == "ContainerCreating" || phase == "" {
		return nil
	}

	return fmt.Errorf("pod is %s (expected networking issue)", phase)
}

func (l *CNIFailureRecoveryLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "cni-test",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("checking pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod not running (phase: %s)", strings.TrimSpace(output))
	}

	return nil
}

func (l *CNIFailureRecoveryLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check CNI pods in kube-system",
			Command:     "kubectl get pods -n kube-system -l k8s-app=calico-node -o wide",
			Notes:       "Check which node the CNI pod is on",
		},
		{
			Description: "Check CNI configuration on the node",
			Command:     "docker exec <control-plane> ls /etc/cni/net.d/",
			Notes:       "Configuration files may be missing",
		},
		{
			Description: "Fix: Restart CNI pods to restore config",
			Command:     "kubectl delete pod -n kube-system -l k8s-app=calico-node",
			Notes:       "Deleting CNI pods forces kubelet to restart them",
		},
		{
			Description: "Verify pod networking works",
			Command:     "kubectl get pod cni-test -o wide",
			Notes:       "Pod should move to Running state",
		},
	}
}
