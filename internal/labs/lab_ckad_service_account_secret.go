package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADServiceAccountSecretLab{})
}

type CKADServiceAccountSecretLab struct {
	BaseLab
}

func (l *CKADServiceAccountSecretLab) ID() string {
	return "ckad_service_account_secret"
}

func (l *CKADServiceAccountSecretLab) Title() string {
	return "Create ServiceAccount with Token Secret"
}

func (l *CKADServiceAccountSecretLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADServiceAccountSecretLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADServiceAccountSecretLab) Cert() Cert             { return CertCKAD }
func (l *CKADServiceAccountSecretLab) DomainWeight() int      { return 25 }
func (l *CKADServiceAccountSecretLab) EstimatedTime() int     { return 15 }
func (l *CKADServiceAccountSecretLab) Tags() []string {
	return []string{"service-account", "token", "authentication"}
}

func (l *CKADServiceAccountSecretLab) Description() string {
	return `An application needs to authenticate to the Kubernetes API using a
ServiceAccount token. Create a ServiceAccount with an associated
token Secret.

Your task: Create the ServiceAccount and manually create a token Secret for it.`
}

func (l *CKADServiceAccountSecretLab) Hints() []string {
	return []string{
		"Create the ServiceAccount first",
		"Manually create a Secret of type kubernetes.io/service-account-token",
		"Annotate the secret with the ServiceAccount name",
	}
}

func (l *CKADServiceAccountSecretLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADServiceAccountSecretLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADServiceAccountSecretLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "secrets",
		"-o", "jsonpath={.items[?(@.type==\"kubernetes.io/service-account-token\")].metadata.annotations}")
	if err != nil {
		return fmt.Errorf("failed to get secrets: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no service-account-token secrets found")
	}
	return nil
}

func (l *CKADServiceAccountSecretLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create ServiceAccount", Command: "kubectl create serviceaccount app-sa"},
		{Description: "Create token secret", Command: "Create Secret with type kubernetes.io/service-account-token and annotation kubernetes.io/service-account.name: app-sa"},
		{Description: "Verify", Command: "kubectl get secrets -o yaml | grep service-account-token"},
	}
}
