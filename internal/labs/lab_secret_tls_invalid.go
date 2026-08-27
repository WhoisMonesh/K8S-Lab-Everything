package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&SecretTLSInvalidLab{})
}

type SecretTLSInvalidLab struct {
	BaseLab
}

func (l *SecretTLSInvalidLab) ID() string {
	return "secret_tls_invalid"
}

func (l *SecretTLSInvalidLab) Title() string {
	return "Secret TLS Type Invalid"
}

func (l *SecretTLSInvalidLab) Category() Category {
	return CategoryWorkloads
}

func (l *SecretTLSInvalidLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *SecretTLSInvalidLab) Description() string {
	return `A Secret 'web-tls' is configured for TLS but uses the wrong type.
The Secret type is 'Opaque' instead of 'kubernetes.io/tls', which causes
ingress TLS termination to fail.

Your task: Fix the Secret type to be a proper TLS secret.`
}

func (l *SecretTLSInvalidLab) Hints() []string {
	return []string{
		"Check the Secret type",
		"TLS secrets must use type 'kubernetes.io/tls'",
		"The Secret must have tls.crt and tls.key data fields",
	}
}

func (l *SecretTLSInvalidLab) EstimatedTime() int {
	return 10
}

func (l *SecretTLSInvalidLab) Tags() []string {
	return []string{"secret", "tls", "ingress", "troubleshooting"}
}

func (l *SecretTLSInvalidLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SecretTLSInvalidLab) Break(ctx context.Context, kubeconfigPath string) error {
	secret := `apiVersion: v1
kind: Secret
metadata:
  name: web-tls
  namespace: default
type: Opaque
data:
  tls.crt: LS0tLS1CRUdJTi...fake
  tls.key: LS0tLS1CRUdJTi...fake
`
	if err := kubectlApply(ctx, kubeconfigPath, secret); err != nil {
		return fmt.Errorf("creating Secret: %w", err)
	}

	ingress := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: web-ingress
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - app.example.com
    secretName: web-tls
  rules:
  - host: app.example.com
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
	if err := kubectlApply(ctx, kubeconfigPath, ingress); err != nil {
		return fmt.Errorf("creating ingress: %w", err)
	}

	return nil
}

func (l *SecretTLSInvalidLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *SecretTLSInvalidLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "secret", "web-tls",
		"-o", "jsonpath={.type}")
	if err != nil {
		return fmt.Errorf("failed to check secret: %w", err)
	}

	if strings.TrimSpace(output) != "kubernetes.io/tls" {
		return fmt.Errorf("secret type is not kubernetes.io/tls (got: %s)", output)
	}

	return nil
}

func (l *SecretTLSInvalidLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check Secret type",
			Command:     "kubectl get secret web-tls -o jsonpath='{.type}'",
			Notes:       "Type is 'Opaque' instead of 'kubernetes.io/tls'",
		},
		{
			Description: "Fix the Secret type",
			Command:     "kubectl patch secret web-tls --type='json' -p='[{\"op\":\"replace\",\"path\":\"/type\",\"value\":\"kubernetes.io/tls\"}]'",
			Notes:       "Change type to kubernetes.io/tls",
		},
		{
			Description: "Verify Secret type",
			Command:     "kubectl get secret web-tls -o jsonpath='{.type}'",
			Notes:       "Type should now be kubernetes.io/tls",
		},
	}
}
