package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSIngressTLSMandatoryLab{})
}

type CKSIngressTLSMandatoryLab struct {
	BaseLab
}

func (l *CKSIngressTLSMandatoryLab) ID() string             { return "cks_ingress_tls_mandatory" }
func (l *CKSIngressTLSMandatoryLab) Title() string          { return "Enforce TLS on All Ingress" }
func (l *CKSIngressTLSMandatoryLab) Category() Category     { return CategoryClusterSetupCKS }
func (l *CKSIngressTLSMandatoryLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSIngressTLSMandatoryLab) EstimatedTime() int     { return 20 }
func (l *CKSIngressTLSMandatoryLab) Cert() Cert             { return CertCKS }
func (l *CKSIngressTLSMandatoryLab) DomainWeight() int      { return 15 }
func (l *CKSIngressTLSMandatoryLab) Tags() []string {
	return []string{"cks", "ingress", "tls", "ssl", "security"}
}

func (l *CKSIngressTLSMandatoryLab) Description() string {
	return `An Ingress resource in namespace 'web-secure' is configured without TLS.
All traffic to this service is transmitted in plaintext.

Your task: Create a TLS Secret named 'web-tls' with a self-signed certificate
and configure the Ingress to use TLS termination on the host 'secure.example.com'.`
}

func (l *CKSIngressTLSMandatoryLab) Hints() []string {
	return []string{
		"Generate a self-signed certificate using openssl",
		"Create a TLS secret with kubectl create secret tls",
		"Add tls section to the Ingress spec",
	}
}

func (l *CKSIngressTLSMandatoryLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSIngressTLSMandatoryLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: web-secure
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	ingress := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: web-ingress
  namespace: web-secure
spec:
  rules:
  - host: secure.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-svc
            port:
              number: 80
`
	return kubectlApply(ctx, kubeconfigPath, ingress)
}

func (l *CKSIngressTLSMandatoryLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "web-ingress", "-n", "web-secure",
		"-o", "jsonpath={.spec.tls}")
	if err != nil {
		return fmt.Errorf("failed to check ingress: %w", err)
	}
	if strings.TrimSpace(output) == "" || output == "null" {
		return fmt.Errorf("TLS not configured on ingress")
	}

	secret, err := kubectl(ctx, kubeconfigPath, "get", "secret", "web-tls", "-n", "web-secure")
	if err != nil {
		return fmt.Errorf("TLS secret not found: %w", err)
	}
	if !strings.Contains(secret, "kubernetes.io/tls") {
		return fmt.Errorf("secret is not a TLS secret")
	}
	return nil
}

func (l *CKSIngressTLSMandatoryLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Generate self-signed certificate", Command: "openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj '/CN=secure.example.com'"},
		{Description: "Create TLS secret", Command: "kubectl create secret tls web-tls --cert=tls.crt --key=tls.key -n web-secure"},
		{Description: "Update ingress with TLS", Command: "kubectl patch ingress web-ingress -n web-secure --type='json' -p='[{\"op\": \"add\", \"path\": \"/spec/tls\", \"value\": [{\"hosts\": [\"secure.example.com\"], \"secretName\": \"web-tls\"}]}'"},
	}
}
