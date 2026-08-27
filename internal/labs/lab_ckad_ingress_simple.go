package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADIngressSimpleLab{})
}

type CKADIngressSimpleLab struct {
	BaseLab
}

func (l *CKADIngressSimpleLab) ID() string             { return "ckad_ingress_simple" }
func (l *CKADIngressSimpleLab) Title() string          { return "Create Simple Ingress" }
func (l *CKADIngressSimpleLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADIngressSimpleLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADIngressSimpleLab) Cert() Cert             { return CertCKAD }
func (l *CKADIngressSimpleLab) DomainWeight() int      { return 20 }
func (l *CKADIngressSimpleLab) EstimatedTime() int     { return 20 }
func (l *CKADIngressSimpleLab) Tags() []string {
	return []string{"ingress", "http", "routing"}
}

func (l *CKADIngressSimpleLab) Description() string {
	return `A web application needs to be exposed via Ingress with path-based routing.
Create an Ingress resource that routes traffic to the backend service.

Your task: Create an Ingress resource for the application.`
}

func (l *CKADIngressSimpleLab) Hints() []string {
	return []string{
		"Use networking.k8s.io/v1 Ingress API",
		"Define rules with host and paths",
		"Configure the backend service and port",
	}
}

func (l *CKADIngressSimpleLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADIngressSimpleLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	svc := `apiVersion: v1
kind: Service
metadata:
  name: web
spec:
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 80`
	return kubectlApply(ctx, kubeconfigPath, svc)
}

func (l *CKADIngressSimpleLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "web-ingress",
		"-o", "jsonpath={.spec.rules[*].host}")
	if err != nil {
		return fmt.Errorf("failed to get ingress: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no Ingress found")
	}
	return nil
}

func (l *CKADIngressSimpleLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create Ingress", Command: "Create Ingress with path / pointing to service web port 80"},
		{Description: "Verify", Command: "kubectl get ingress web-ingress"},
	}
}
