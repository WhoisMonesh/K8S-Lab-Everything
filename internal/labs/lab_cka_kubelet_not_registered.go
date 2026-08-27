package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&KubeletNotRegisteredLab{})
}

type KubeletNotRegisteredLab struct {
	BaseLab
}

func (l *KubeletNotRegisteredLab) ID() string { return "cka_kubelet_not_registered" }
func (l *KubeletNotRegisteredLab) Title() string {
	return "Debug Kubelet Registration Issues"
}
func (l *KubeletNotRegisteredLab) Category() Category     { return CategoryTroubleshooting }
func (l *KubeletNotRegisteredLab) Difficulty() Difficulty { return DifficultyHard }
func (l *KubeletNotRegisteredLab) EstimatedTime() int     { return 25 }
func (l *KubeletNotRegisteredLab) Tags() []string {
	return []string{"kubelet", "registration", "node", "troubleshooting"}
}
func (l *KubeletNotRegisteredLab) Cert() Cert        { return CertCKA }
func (l *KubeletNotRegisteredLab) DomainWeight() int { return 30 }

func (l *KubeletNotRegisteredLab) Description() string {
	return `A kubelet on a worker node is not registered with the API server.
The node is missing from kubectl get nodes. Diagnose and fix the
kubelet registration issue.`
}

func (l *KubeletNotRegisteredLab) Hints() []string {
	return []string{
		"Check kubelet logs on the worker node",
		"Verify kubelet configuration",
		"Check certificates are present",
	}
}

func (l *KubeletNotRegisteredLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeletNotRegisteredLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *KubeletNotRegisteredLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o", "name")
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("worker node not registered")
	}
	return nil
}

func (l *KubeletNotRegisteredLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check nodes", Command: "kubectl get nodes"},
		{Description: "Check kubelet logs", Command: "ssh <node> journalctl -u kubelet -f"},
		{Description: "Check kubelet config", Command: "ssh <node> cat /var/lib/kubelet/config.yaml"},
		{Description: "Restart kubelet", Command: "ssh <node> systemctl restart kubelet"},
	}
}
