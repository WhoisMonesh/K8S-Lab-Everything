package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADServiceAccountCreateLab{})
}

type CKADServiceAccountCreateLab struct {
	BaseLab
}

func (l *CKADServiceAccountCreateLab) ID() string {
	return "ckad_service_account_create"
}

func (l *CKADServiceAccountCreateLab) Title() string          { return "Create ServiceAccount" }
func (l *CKADServiceAccountCreateLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADServiceAccountCreateLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADServiceAccountCreateLab) Cert() Cert             { return CertCKAD }
func (l *CKADServiceAccountCreateLab) DomainWeight() int      { return 25 }
func (l *CKADServiceAccountCreateLab) EstimatedTime() int     { return 10 }
func (l *CKADServiceAccountCreateLab) Tags() []string {
	return []string{"service-account", "rbac", "security"}
}

func (l *CKADServiceAccountCreateLab) Description() string {
	return `An application needs a dedicated ServiceAccount with specific permissions.
Create a ServiceAccount named 'app-sa' in the 'app-ns' namespace.

Your task: Create the ServiceAccount and configure a pod to use it.`
}

func (l *CKADServiceAccountCreateLab) Hints() []string {
	return []string{
		"Use kubectl create serviceaccount",
		"Specify the namespace with -n",
		"Set serviceAccountName in the pod spec",
	}
}

func (l *CKADServiceAccountCreateLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADServiceAccountCreateLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: app-ns`
	return kubectlApply(ctx, kubeconfigPath, ns)
}

func (l *CKADServiceAccountCreateLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "serviceaccount", "app-sa",
		"-n", "app-ns", "-o", "jsonpath={.metadata.name}")
	if err != nil {
		return fmt.Errorf("failed to get serviceaccount: %w", err)
	}
	if strings.TrimSpace(output) != "app-sa" {
		return fmt.Errorf("serviceaccount 'app-sa' not found")
	}
	return nil
}

func (l *CKADServiceAccountCreateLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create namespace", Command: "kubectl create namespace app-ns"},
		{Description: "Create ServiceAccount", Command: "kubectl create serviceaccount app-sa -n app-ns"},
		{Description: "Verify", Command: "kubectl get serviceaccount app-sa -n app-ns"},
	}
}
