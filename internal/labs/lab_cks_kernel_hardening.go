package labs

import (
	"context"
)

func init() {
	Register(&CKSKernelHardeningLab{})
}

type CKSKernelHardeningLab struct {
	BaseLab
}

func (l *CKSKernelHardeningLab) ID() string             { return "cks_kernel_hardening" }
func (l *CKSKernelHardeningLab) Title() string          { return "Configure Kernel Hardening Parameters" }
func (l *CKSKernelHardeningLab) Category() Category     { return CategorySystemHardening }
func (l *CKSKernelHardeningLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSKernelHardeningLab) EstimatedTime() int     { return 25 }
func (l *CKSKernelHardeningLab) Cert() Cert             { return CertCKS }
func (l *CKSKernelHardeningLab) DomainWeight() int      { return 10 }
func (l *CKSKernelHardeningLab) Tags() []string {
	return []string{"cks", "kernel", "sysctl", "hardening"}
}

func (l *CKSKernelHardeningLab) Description() string {
	return `The kernel parameters on worker nodes are not hardened. Several security-relevant
sysctl settings are at their default (insecure) values.

Your task: Configure the following kernel parameters:
- net.ipv4.ip_forward = 1 (required for kube-proxy)
- net.bridge.bridge-nf-call-iptables = 1 (required for CNI)
- net.ipv4.conf.all.forwarding = 1 (for IPv4 forwarding)`
}

func (l *CKSKernelHardeningLab) Hints() []string {
	return []string{
		"Use sysctl to set kernel parameters",
		"Create /etc/sysctl.d/ files for persistent configuration",
		"Apply with sysctl --system",
	}
}

func (l *CKSKernelHardeningLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSKernelHardeningLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSKernelHardeningLab) Verify(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSKernelHardeningLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current sysctl settings", Command: "sysctl net.ipv4.ip_forward net.bridge.bridge-nf-call-iptables"},
		{Description: "Create sysctl config", Command: "sudo tee /etc/sysctl.d/99-kubelet.conf <<EOF\nnet.ipv4.ip_forward=1\nnet.bridge.bridge-nf-call-iptables=1\nnet.ipv4.conf.all.forwarding=1\nEOF"},
		{Description: "Apply sysctl settings", Command: "sudo sysctl --system"},
		{Description: "Verify settings", Command: "sysctl net.ipv4.ip_forward"},
	}
}
