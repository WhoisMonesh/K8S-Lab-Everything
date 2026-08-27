package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSImageScanningTrivyLab{})
}

type CKSImageScanningTrivyLab struct {
	BaseLab
}

func (l *CKSImageScanningTrivyLab) ID() string             { return "cks_image_scanning_trivy" }
func (l *CKSImageScanningTrivyLab) Title() string          { return "Scan Images with Trivy" }
func (l *CKSImageScanningTrivyLab) Category() Category     { return CategorySupplyChain }
func (l *CKSImageScanningTrivyLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSImageScanningTrivyLab) EstimatedTime() int     { return 25 }
func (l *CKSImageScanningTrivyLab) Cert() Cert             { return CertCKS }
func (l *CKSImageScanningTrivyLab) DomainWeight() int      { return 20 }
func (l *CKSImageScanningTrivyLab) Tags() []string {
	return []string{"cks", "trivy", "image-scanning", "supply-chain"}
}

func (l *CKSImageScanningTrivyLab) Description() string {
	return `Container images in the cluster have not been scanned for vulnerabilities.
Deploy the Trivy operator to continuously scan images in the cluster.

Your task: Install the Trivy operator in the 'trivy-system' namespace
and verify it is scanning workloads.`
}

func (l *CKSImageScanningTrivyLab) Hints() []string {
	return []string{
		"Use Helm to install the Trivy operator",
		"helm repo add aquasecurity https://aquasecurity.github.io/helm-charts/",
		"helm install trivy-operator aquasecurity/trivy-operator -n trivy-system --create-namespace",
	}
}

func (l *CKSImageScanningTrivyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSImageScanningTrivyLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSImageScanningTrivyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "trivy-system", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get trivy pods: %w", err)
	}
	if strings.Contains(output, "trivy-operator") {
		return nil
	}
	return fmt.Errorf("trivy operator not deployed")
}

func (l *CKSImageScanningTrivyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Add Trivy Helm repo", Command: "helm repo add aquasecurity https://aquasecurity.github.io/helm-charts/ && helm repo update"},
		{Description: "Install Trivy operator", Command: "helm install trivy-operator aquasecurity/trivy-operator -n trivy-system --create-namespace"},
		{Description: "Verify installation", Command: "kubectl get pods -n trivy-system"},
	}
}
