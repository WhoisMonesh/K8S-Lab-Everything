package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&SecretEncryptionDisabled{})
}

type SecretEncryptionDisabled struct {
	BaseLab
}

func (l *SecretEncryptionDisabled) ID() string             { return "secret_encryption_disabled" }
func (l *SecretEncryptionDisabled) Title() string          { return "Secret Encryption at Rest Disabled" }
func (l *SecretEncryptionDisabled) Category() Category     { return CategorySecurity }
func (l *SecretEncryptionDisabled) Difficulty() Difficulty { return DifficultyHard }
func (l *SecretEncryptionDisabled) EstimatedTime() int     { return 25 }
func (l *SecretEncryptionDisabled) Tags() []string {
	return []string{"security", "encryption", "secrets"}
}

func (l *SecretEncryptionDisabled) Description() string {
	return `Secrets are stored in etcd without encryption. Configure encryption at rest for Secrets.`
}

func (l *SecretEncryptionDisabled) Hints() []string {
	return []string{
		"Check etcd encryption configuration",
		"Create an EncryptionConfiguration resource",
		"Restart the API server after changes",
	}
}

func (l *SecretEncryptionDisabled) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SecretEncryptionDisabled) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *SecretEncryptionDisabled) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system", "-l", "component=kube-apiserver",
		"-o", "jsonpath={.items[0].spec.containers[0].args}")
	if err != nil {
		return err
	}
	if containsAny(output, "encryption-provider-config") {
		return nil
	}
	return fmt.Errorf("encryption not configured")
}

func (l *SecretEncryptionDisabled) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create encryption config", Command: "sudo tee /etc/kubernetes/encryption-config.yaml <<EOF\napiVersion: apiserver.config.k8s.io/v1\nkind: EncryptionConfiguration\nresources:\n  - resources:\n      - secrets\n    providers:\n      - aescbc:\n          keys:\n            - name: key1\n              secret: <base64-encoded-key>\n      - identity: {}\nEOF"},
		{Description: "Add encryption flag", Command: "Edit kube-apiserver.yaml and add --encryption-provider-config=/etc/kubernetes/encryption-config.yaml"},
		{Description: "Restart API server", Command: "sudo mv /etc/kubernetes/manifests/kube-apiserver.yaml /tmp/ && sleep 10 && sudo mv /tmp/kube-apiserver.yaml /etc/kubernetes/manifests/"},
	}
}
