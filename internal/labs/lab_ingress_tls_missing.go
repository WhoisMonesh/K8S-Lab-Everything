package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&IngressTLSMissing{})
}

type IngressTLSMissing struct {
	BaseLab
}

func (l *IngressTLSMissing) ID() string             { return "ingress_tls_missing" }
func (l *IngressTLSMissing) Title() string          { return "Ingress TLS Secret Missing" }
func (l *IngressTLSMissing) Category() Category     { return CategoryNetworking }
func (l *IngressTLSMissing) Difficulty() Difficulty { return DifficultyMedium }
func (l *IngressTLSMissing) EstimatedTime() int     { return 20 }
func (l *IngressTLSMissing) Tags() []string         { return []string{"ingress", "tls", "certificates"} }

func (l *IngressTLSMissing) Description() string {
	return `An Ingress resource is configured for TLS but the referenced Secret does not exist.
Create the TLS Secret and fix the Ingress configuration.`
}

func (l *IngressTLSMissing) Hints() []string {
	return []string{
		"Check the Ingress TLS configuration",
		"Look at the secretName referenced in tls section",
		"Create the TLS Secret with correct keys",
	}
}

func (l *IngressTLSMissing) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *IngressTLSMissing) Break(ctx context.Context, kubeconfigPath string) error {
	ingress := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: secure-app
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - secure.example.com
    secretName: tls-secret
  rules:
  - host: secure.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: secure-app
            port:
              number: 80
---
apiVersion: v1
kind: Service
metadata:
  name: secure-app
spec:
  selector:
    app: secure-app
  ports:
  - port: 80
    targetPort: 80`
	return kubectlApply(ctx, kubeconfigPath, ingress)
}

func (l *IngressTLSMissing) Verify(ctx context.Context, kubeconfigPath string) error {
	_, err := kubectl(ctx, kubeconfigPath, "get", "secret", "tls-secret")
	if err != nil {
		return fmt.Errorf("TLS secret tls-secret not found")
	}
	return nil
}

func (l *IngressTLSMissing) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check Ingress TLS config", Command: "kubectl describe ingress secure-app"},
		{Description: "Check if secret exists", Command: "kubectl get secret tls-secret"},
		{Description: "Create TLS secret", Command: "kubectl create secret tls tls-secret --cert=tls.crt --key=tls.key"},
	}
}
