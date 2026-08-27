package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSFalcoRuntimeLab{})
}

type CKSFalcoRuntimeLab struct {
	BaseLab
}

func (l *CKSFalcoRuntimeLab) ID() string             { return "cks_falco_runtime" }
func (l *CKSFalcoRuntimeLab) Title() string          { return "Deploy Falco Runtime Security" }
func (l *CKSFalcoRuntimeLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSFalcoRuntimeLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSFalcoRuntimeLab) EstimatedTime() int     { return 25 }
func (l *CKSFalcoRuntimeLab) Cert() Cert             { return CertCKS }
func (l *CKSFalcoRuntimeLab) DomainWeight() int      { return 20 }
func (l *CKSFalcoRuntimeLab) Tags() []string {
	return []string{"cks", "falco", "runtime-security", "monitoring"}
}

func (l *CKSFalcoRuntimeLab) Description() string {
	return `The cluster has no runtime security monitoring. Malicious activity inside
containers goes undetected.

Your task: Deploy Falco as a DaemonSet in the 'falco' namespace to monitor
runtime behavior and detect suspicious system calls.`
}

func (l *CKSFalcoRuntimeLab) Hints() []string {
	return []string{
		"Use Helm to install Falco",
		"helm repo add falcosecurity https://falcosecurity.github.io/charts",
		"helm install falco falcosecurity/falco -n falco --create-namespace",
	}
}

func (l *CKSFalcoRuntimeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSFalcoRuntimeLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSFalcoRuntimeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "falco", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get falco pods: %w", err)
	}
	if strings.Contains(output, "falco") {
		return nil
	}
	return fmt.Errorf("falco not deployed")
}

func (l *CKSFalcoRuntimeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Add Falco Helm repo", Command: "helm repo add falcosecurity https://falcosecurity.github.io/charts/ && helm repo update"},
		{Description: "Install Falco", Command: "helm install falco falcosecurity/falco -n falco --create-namespace"},
		{Description: "Verify installation", Command: "kubectl get pods -n falco"},
	}
}
