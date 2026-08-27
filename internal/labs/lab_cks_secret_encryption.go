package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSSecretEncryptionLab{})
}

type CKSSecretEncryptionLab struct {
	BaseLab
}

func (l *CKSSecretEncryptionLab) ID() string             { return "cks_secret_encryption" }
func (l *CKSSecretEncryptionLab) Title() string          { return "Encrypt Secrets at Rest" }
func (l *CKSSecretEncryptionLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSSecretEncryptionLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSSecretEncryptionLab) EstimatedTime() int     { return 30 }
func (l *CKSSecretEncryptionLab) Cert() Cert             { return CertCKS }
func (l *CKSSecretEncryptionLab) DomainWeight() int      { return 20 }
func (l *CKSSecretEncryptionLab) Tags() []string {
	return []string{"cks", "secrets", "encryption", "etcd"}
}

func (l *CKSSecretEncryptionLab) Description() string {
	return `Secrets are stored in etcd without encryption at rest. This means anyone
with access to the etcd data can read all Secret values.

Your task: Configure encryption at rest for Secrets using an
EncryptionConfiguration with aescbc provider.`
}

func (l *CKSSecretEncryptionLab) Hints() []string {
	return []string{
		"Generate an encryption key with head -c 32 /dev/urandom | base64",
		"Create EncryptionConfiguration YAML",
		"Add --encryption-provider-config flag to kube-apiserver",
	}
}

func (l *CKSSecretEncryptionLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSSecretEncryptionLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSSecretEncryptionLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-apiserver", "-o", "jsonpath={.items[0].spec.containers[0].args}")
	if err != nil {
		return fmt.Errorf("failed to get apiserver args: %w", err)
	}
	if strings.Contains(output, "encryption-provider-config") {
		return nil
	}
	return fmt.Errorf("encryption provider config not configured")
}

func (l *CKSSecretEncryptionLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Generate encryption key", Command: "head -c 32 /dev/urandom | base64"},
		{Description: "Create encryption config", Command: "sudo tee /etc/kubernetes/encryption-config.yaml <<EOF\napiVersion: apiserver.config.k8s.io/v1\nkind: EncryptionConfiguration\nresources:\n  - resources:\n      - secrets\n    providers:\n      - aescbc:\n          keys:\n            - name: key1\n              secret: <base64-key>\n      - identity: {}\nEOF"},
		{Description: "Add to API server manifest", Command: "Edit /etc/kubernetes/manifests/kube-apiserver.yaml and add --encryption-provider-config=/etc/kubernetes/encryption-config.yaml"},
		{Description: "Restart API server", Command: "sudo mv /etc/kubernetes/manifests/kube-apiserver.yaml /tmp/ && sleep 10 && sudo mv /tmp/kube-apiserver.yaml /etc/kubernetes/manifests/"},
	}
}
