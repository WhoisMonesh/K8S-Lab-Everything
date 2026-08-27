package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADSecretGenericLab{})
}

type CKADSecretGenericLab struct {
	BaseLab
}

func (l *CKADSecretGenericLab) ID() string             { return "ckad_secret_generic" }
func (l *CKADSecretGenericLab) Title() string          { return "Create Generic Secret" }
func (l *CKADSecretGenericLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADSecretGenericLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADSecretGenericLab) Cert() Cert             { return CertCKAD }
func (l *CKADSecretGenericLab) DomainWeight() int      { return 25 }
func (l *CKADSecretGenericLab) EstimatedTime() int     { return 10 }
func (l *CKADSecretGenericLab) Tags() []string {
	return []string{"secret", "generic", "configuration"}
}

func (l *CKADSecretGenericLab) Description() string {
	return `An application needs sensitive configuration data. Create a generic Secret
named 'app-secret' with the required key-value pairs.

Your task: Create the generic Secret with the specified data.`
}

func (l *CKADSecretGenericLab) Hints() []string {
	return []string{
		"Use kubectl create secret generic",
		"Provide --from-literal for each key-value pair",
		"Values are base64 encoded automatically",
	}
}

func (l *CKADSecretGenericLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADSecretGenericLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADSecretGenericLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "secret", "app-secret",
		"-o", "jsonpath={.type}")
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}
	if strings.TrimSpace(output) != "Opaque" {
		return fmt.Errorf("secret type is not Opaque (current: %s)", output)
	}
	return nil
}

func (l *CKADSecretGenericLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create generic secret", Command: "kubectl create secret generic app-secret --from-literal=username=admin --from-literal=password=s3cr3t"},
		{Description: "Verify", Command: "kubectl get secret app-secret -o yaml"},
	}
}
