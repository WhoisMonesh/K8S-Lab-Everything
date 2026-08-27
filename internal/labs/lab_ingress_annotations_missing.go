package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&IngressAnnotationsMissingLab{})
}

type IngressAnnotationsMissingLab struct {
	BaseLab
}

func (l *IngressAnnotationsMissingLab) ID() string {
	return "ingress_annotations_missing"
}

func (l *IngressAnnotationsMissingLab) Title() string {
	return "Ingress Annotations Required but Missing"
}

func (l *IngressAnnotationsMissingLab) Category() Category {
	return CategoryNetworking
}

func (l *IngressAnnotationsMissingLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *IngressAnnotationsMissingLab) Description() string {
	return `An Ingress 'ssl-app' requires TLS termination but is missing the
nginx.ingress.kubernetes.io/ssl-redirect annotation. Without it,
HTTP traffic is not redirected to HTTPS.

Your task: Add the required annotation to enable SSL redirect.`
}

func (l *IngressAnnotationsMissingLab) Hints() []string {
	return []string{
		"Check the Ingress annotations",
		"The ssl-redirect annotation is needed for HTTPS redirect",
		"Add nginx.ingress.kubernetes.io/ssl-redirect: \"true\"",
	}
}

func (l *IngressAnnotationsMissingLab) EstimatedTime() int {
	return 10
}

func (l *IngressAnnotationsMissingLab) Tags() []string {
	return []string{"ingress", "annotations", "ssl", "tls", "networking"}
}

func (l *IngressAnnotationsMissingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *IngressAnnotationsMissingLab) Break(ctx context.Context, kubeconfigPath string) error {
	ingress := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ssl-app
  namespace: default
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
            name: ssl-svc
            port:
              number: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, ingress); err != nil {
		return fmt.Errorf("creating ingress: %w", err)
	}

	return nil
}

func (l *IngressAnnotationsMissingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *IngressAnnotationsMissingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "ssl-app",
		"-o", "jsonpath={.metadata.annotations.nginx\\.ingress\\.kubernetes\\.io/ssl-redirect}")
	if err != nil {
		return fmt.Errorf("failed to check ingress: %w", err)
	}

	if strings.TrimSpace(output) != "true" {
		return fmt.Errorf("ssl-redirect annotation not set to true")
	}

	return nil
}

func (l *IngressAnnotationsMissingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check Ingress annotations",
			Command:     "kubectl get ingress ssl-app -o yaml | grep annotations -A 5",
			Notes:       "No ssl-redirect annotation present",
		},
		{
			Description: "Add SSL redirect annotation",
			Command:     "kubectl annotate ingress ssl-app nginx.ingress.kubernetes.io/ssl-redirect=true",
			Notes:       "This enables automatic HTTP to HTTPS redirect",
		},
		{
			Description: "Verify annotation",
			Command:     "kubectl get ingress ssl-app -o yaml | grep ssl-redirect",
			Notes:       "Should now show ssl-redirect: \"true\"",
		},
	}
}
