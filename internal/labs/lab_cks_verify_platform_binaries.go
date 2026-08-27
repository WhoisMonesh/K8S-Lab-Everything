package labs

import (
	"context"
)

func init() {
	Register(&CKSVerifyPlatformBinariesLab{})
}

type CKSVerifyPlatformBinariesLab struct {
	BaseLab
}

func (l *CKSVerifyPlatformBinariesLab) ID() string             { return "cks_verify_platform_binaries" }
func (l *CKSVerifyPlatformBinariesLab) Title() string          { return "Verify Platform Binary Checksums" }
func (l *CKSVerifyPlatformBinariesLab) Category() Category     { return CategoryClusterSetupCKS }
func (l *CKSVerifyPlatformBinariesLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSVerifyPlatformBinariesLab) EstimatedTime() int     { return 15 }
func (l *CKSVerifyPlatformBinariesLab) Cert() Cert             { return CertCKS }
func (l *CKSVerifyPlatformBinariesLab) DomainWeight() int      { return 15 }
func (l *CKSVerifyPlatformBinariesLab) Tags() []string {
	return []string{"cks", "binaries", "checksum", "supply-chain"}
}

func (l *CKSVerifyPlatformBinariesLab) Description() string {
	return `The Kubernetes platform binaries (kubeadm, kubelet, kubectl) on worker nodes
have not been verified against official checksums.

Your task: Verify the checksum of the kubelet binary on the first worker node
against the official release checksum. Document the verification result.`
}

func (l *CKSVerifyPlatformBinariesLab) Hints() []string {
	return []string{
		"Download the official checksum file for the Kubernetes version",
		"Use sha256sum to compute the local binary checksum",
		"Compare the two checksums",
	}
}

func (l *CKSVerifyPlatformBinariesLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSVerifyPlatformBinariesLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSVerifyPlatformBinariesLab) Verify(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSVerifyPlatformBinariesLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Get Kubernetes version", Command: "kubectl version --short | grep Server"},
		{Description: "Check kubelet binary path", Command: "which kubelet"},
		{Description: "Compute local checksum", Command: "sha256sum $(which kubelet)"},
		{Description: "Download official checksum", Command: "curl -LO https://dl.k8s.io/release/stable.txt && curl -LO https://dl.k8s.io/release/$(cat stable.txt)/bin/linux/amd64/kubelet.sha256"},
		{Description: "Compare checksums", Command: "cat kubelet.sha256 && echo 'Compare with local sha256sum output'"},
	}
}
