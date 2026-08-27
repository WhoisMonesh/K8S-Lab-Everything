package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADSecretTLSLab{})
}

type CKADSecretTLSLab struct {
	BaseLab
}

func (l *CKADSecretTLSLab) ID() string             { return "ckad_secret_tls" }
func (l *CKADSecretTLSLab) Title() string          { return "Create TLS Secret" }
func (l *CKADSecretTLSLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADSecretTLSLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADSecretTLSLab) Cert() Cert             { return CertCKAD }
func (l *CKADSecretTLSLab) DomainWeight() int      { return 25 }
func (l *CKADSecretTLSLab) EstimatedTime() int     { return 15 }
func (l *CKADSecretTLSLab) Tags() []string {
	return []string{"secret", "tls", "certificates"}
}

func (l *CKADSecretTLSLab) Description() string {
	return `An Ingress resource needs TLS termination. Create a TLS Secret named
'my-tls' with the provided certificate and key files.

Your task: Create the TLS Secret from the certificate files.`
}

func (l *CKADSecretTLSLab) Hints() []string {
	return []string{
		"Use kubectl create secret tls",
		"Provide --cert and --key flags",
		"The secret type must be kubernetes.io/tls",
	}
}

func (l *CKADSecretTLSLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADSecretTLSLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADSecretTLSLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "secret", "my-tls",
		"-o", "jsonpath={.type}")
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}
	if strings.TrimSpace(output) != "kubernetes.io/tls" {
		return fmt.Errorf("secret type is not TLS (current: %s)", output)
	}
	return nil
}

func (l *CKADSecretTLSLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Generate self-signed cert", Command: "openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt"},
		{Description: "Create TLS Secret", Command: "kubectl create secret tls my-tls --cert=tls.crt --key=tls.key"},
		{Description: "Verify", Command: "kubectl get secret my-tls -o yaml"},
	}
}
