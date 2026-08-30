package labs

import (
	"context"
	"fmt"
	"time"
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
	return `The Kubernetes platform binaries on worker nodes have not been verified against
official checksums. Without this verification, tampered binaries could go
undetected.

Your task: On a worker node, compute the SHA-256 checksum of the kubelet binary,
compare it against the official release checksum, and document verification by
creating the marker file /var/lib/cka-platform/checksum.verified containing the
kubelet reference.`
}

func (l *CKSVerifyPlatformBinariesLab) Hints() []string {
	return []string{
		"Enter the worker node shell with docker exec",
		"Compute sha256 of the kubelet binary",
		"Compare with the official checksum and write the marker file",
	}
}

func (l *CKSVerifyPlatformBinariesLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSVerifyPlatformBinariesLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSVerifyPlatformBinariesLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *CKSVerifyPlatformBinariesLab) Verify(ctx context.Context, kubeconfigPath string) error {
	worker, err := getWorkerNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	if _, err := dockerCommand(worker, "test -f /var/lib/cka-platform/checksum.verified && grep -q kubelet /var/lib/cka-platform/checksum.verified"); err != nil {
		return fmt.Errorf("platform binaries not verified")
	}
	return nil
}

func (l *CKSVerifyPlatformBinariesLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Enter the worker node shell (kind has no SSH)", Command: "docker exec -it <worker> bash"},
		{Description: "Compute and compare the kubelet checksum, then create the marker file", Command: "mkdir -p /var/lib/cka-platform && sha256sum $(which kubelet) > /var/lib/cka-platform/checksum.verified && cat /var/lib/cka-platform/checksum.verified"},
	}
}
