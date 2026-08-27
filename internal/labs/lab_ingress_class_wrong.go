package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&IngressClassWrongLab{})
}

type IngressClassWrongLab struct {
	BaseLab
}

func (l *IngressClassWrongLab) ID() string {
	return "ingress_class_wrong"
}

func (l *IngressClassWrongLab) Title() string {
	return "Ingress Using Wrong Class"
}

func (l *IngressClassWrongLab) Category() Category {
	return CategoryNetworking
}

func (l *IngressClassWrongLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *IngressClassWrongLab) Description() string {
	return `An Ingress 'api-gateway' is using ingressClass 'nginx' but the cluster
only has an IngressClass named 'traefik'. The Ingress is not being
processed by any controller.

Your task: Fix the Ingress to use the correct IngressClass.`
}

func (l *IngressClassWrongLab) Hints() []string {
	return []string{
		"Check available IngressClasses",
		"The Ingress references an IngressClass that doesn't exist",
		"Update the Ingress to use the available IngressClass",
	}
}

func (l *IngressClassWrongLab) EstimatedTime() int {
	return 10
}

func (l *IngressClassWrongLab) Tags() []string {
	return []string{"ingress", "ingress-class", "controller", "networking"}
}

func (l *IngressClassWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *IngressClassWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	ingressClass := `apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: traefik
spec:
  controller: traefik.io/ingress-controller
`
	if err := kubectlApply(ctx, kubeconfigPath, ingressClass); err != nil {
		return fmt.Errorf("creating IngressClass: %w", err)
	}

	ingress := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-gateway
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  ingressClassName: nginx
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-svc
            port:
              number: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, ingress); err != nil {
		return fmt.Errorf("creating ingress: %w", err)
	}

	return nil
}

func (l *IngressClassWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *IngressClassWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "api-gateway",
		"-o", "jsonpath={.spec.ingressClassName}")
	if err != nil {
		return fmt.Errorf("failed to check ingress: %w", err)
	}

	if strings.TrimSpace(output) == "nginx" {
		return fmt.Errorf("ingressClassName is still nginx")
	}

	return nil
}

func (l *IngressClassWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check available IngressClasses",
			Command:     "kubectl get ingressclasses",
			Notes:       "Only traefik IngressClass exists",
		},
		{
			Description: "Check Ingress configuration",
			Command:     "kubectl get ingress api-gateway -o yaml | grep ingressClassName",
			Notes:       "IngressClassName is nginx which doesn't exist",
		},
		{
			Description: "Fix IngressClass",
			Command:     "kubectl patch ingress api-gateway --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/ingressClassName\",\"value\":\"traefik\"}]'",
			Notes:       "Change to traefik which is the available IngressClass",
		},
		{
			Description: "Verify Ingress is processed",
			Command:     "kubectl get ingress api-gateway",
			Notes:       " ADDRESS column should now show an IP",
		},
	}
}
