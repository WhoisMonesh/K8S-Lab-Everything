package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSSecretExternalSecretsLab{})
}

type CKSSecretExternalSecretsLab struct {
	BaseLab
}

func (l *CKSSecretExternalSecretsLab) ID() string             { return "cks_secret_external_secrets" }
func (l *CKSSecretExternalSecretsLab) Title() string          { return "Use External Secrets Provider" }
func (l *CKSSecretExternalSecretsLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSSecretExternalSecretsLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSSecretExternalSecretsLab) EstimatedTime() int     { return 30 }
func (l *CKSSecretExternalSecretsLab) Cert() Cert             { return CertCKS }
func (l *CKSSecretExternalSecretsLab) DomainWeight() int      { return 20 }
func (l *CKSSecretExternalSecretsLab) Tags() []string {
	return []string{"cks", "secrets", "external-secrets", "vault"}
}

func (l *CKSSecretExternalSecretsLab) Description() string {
	return `Sensitive credentials are stored directly in Kubernetes Secrets.
Instead, secrets should be fetched from an external secrets manager.

Your task: Create a SecretProviderClass resource that references
an external secrets store (aws-secrets-manager) for the secret 'db-password'
in region 'us-east-1'.`
}

func (l *CKSSecretExternalSecretsLab) Hints() []string {
	return []string{
		"Use the CSI SecretProviderClass CRD",
		"Reference the aws-secrets-manager provider",
		"Specify the secret ARN or name in the objects section",
	}
}

func (l *CKSSecretExternalSecretsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSSecretExternalSecretsLab) Break(ctx context.Context, kubeconfigPath string) error {
	secret := `apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
  namespace: default
type: Opaque
stringData:
  password: supersecret123
`
	return kubectlApply(ctx, kubeconfigPath, secret)
}

func (l *CKSSecretExternalSecretsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "secretproviderclass", "-A", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get secretproviderclass: %w", err)
	}
	if !strings.Contains(output, "aws-secrets-manager") {
		return fmt.Errorf("SecretProviderClass not configured with external provider")
	}
	return nil
}

func (l *CKSSecretExternalSecretsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create SecretProviderClass", Command: `cat <<EOF | kubectl apply -f -
apiVersion: secrets-store.csi.x-k8s.io/v1
kind: SecretProviderClass
metadata:
  name: aws-secrets
spec:
  provider: aws
  parameters:
    objects: |
      - objectName: db-password
        objectType: secretsmanager
        objectAlias: db-password
EOF`},
		{Description: "Verify", Command: "kubectl get secretproviderclass aws-secrets"},
	}
}
