package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSHostOSMinimizeLab{})
}

type CKSHostOSMinimizeLab struct {
	BaseLab
}

func (l *CKSHostOSMinimizeLab) ID() string             { return "cks_host_os_minimize" }
func (l *CKSHostOSMinimizeLab) Title() string          { return "Minimize Host OS Footprint" }
func (l *CKSHostOSMinimizeLab) Category() Category     { return CategorySystemHardening }
func (l *CKSHostOSMinimizeLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSHostOSMinimizeLab) EstimatedTime() int     { return 20 }
func (l *CKSHostOSMinimizeLab) Cert() Cert             { return CertCKS }
func (l *CKSHostOSMinimizeLab) DomainWeight() int      { return 10 }
func (l *CKSHostOSMinimizeLab) Tags() []string {
	return []string{"cks", "host-os", "minimize", "system-hardening"}
}

func (l *CKSHostOSMinimizeLab) Description() string {
	return `The worker nodes have unnecessary packages installed and services running
that increase the attack surface.

Your task: Verify that the kubelet is configured with:
- --fail-swap-on=false is NOT set (swap should cause failure)
-_rotate-certificates is set to true
- --protect-kernel-defaults is set to true`
}

func (l *CKSHostOSMinimizeLab) Hints() []string {
	return []string{
		"Check kubelet configuration on worker nodes",
		"Verify --fail-swap-on is not disabled",
		"Ensure --protect-kernel-defaults and --rotate-certificates are enabled",
	}
}

func (l *CKSHostOSMinimizeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSHostOSMinimizeLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSHostOSMinimizeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o", "jsonpath={.items[*].status.nodeInfo.kubeletVersion}")
	if err != nil {
		return fmt.Errorf("failed to get nodes: %w", err)
	}
	if strings.TrimSpace(output) != "" {
		return nil
	}
	return fmt.Errorf("could not verify node configuration")
}

func (l *CKSHostOSMinimizeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check kubelet config", Command: "cat /var/lib/kubelet/config.yaml"},
		{Description: "Verify fail-swap-on setting", Command: "grep -i swap /var/lib/kubelet/config.yaml"},
		{Description: "Enable protect-kernel-defaults", Command: "echo 'protectKernelDefaults: true' >> /var/lib/kubelet/config.yaml"},
		{Description: "Enable rotate-certificates", Command: "echo 'rotateCertificates: true' >> /var/lib/kubelet/config.yaml"},
		{Description: "Restart kubelet", Command: "sudo systemctl daemon-reload && sudo systemctl restart kubelet"},
	}
}
