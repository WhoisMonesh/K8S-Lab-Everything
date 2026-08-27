package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADServiceNodePortLab{})
}

type CKADServiceNodePortLab struct {
	BaseLab
}

func (l *CKADServiceNodePortLab) ID() string             { return "ckad_service_nodeport" }
func (l *CKADServiceNodePortLab) Title() string          { return "Create NodePort Service" }
func (l *CKADServiceNodePortLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADServiceNodePortLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADServiceNodePortLab) Cert() Cert             { return CertCKAD }
func (l *CKADServiceNodePortLab) DomainWeight() int      { return 20 }
func (l *CKADServiceNodePortLab) EstimatedTime() int     { return 10 }
func (l *CKADServiceNodePortLab) Tags() []string {
	return []string{"service", "nodeport", "external-access"}
}

func (l *CKADServiceNodePortLab) Description() string {
	return `An application needs to be accessible from outside the cluster on a
specific port. Create a NodePort service named 'web-svc'.

Your task: Create the NodePort service exposing port 80.`
}

func (l *CKADServiceNodePortLab) Hints() []string {
	return []string{
		"NodePort exposes the service on each node's IP",
		"Use --type=NodePort",
		"You can specify a port or let Kubernetes assign one",
	}
}

func (l *CKADServiceNodePortLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADServiceNodePortLab) Break(ctx context.Context, kubeconfigPath string) error {
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
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *CKADServiceNodePortLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "web-svc",
		"-o", "jsonpath={.spec.type}")
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}
	if strings.TrimSpace(output) != "NodePort" {
		return fmt.Errorf("service type is not NodePort (current: %s)", output)
	}
	return nil
}

func (l *CKADServiceNodePortLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create service", Command: "kubectl expose deployment web --name=web-svc --port=80 --target-port=80 --type=NodePort"},
		{Description: "Verify", Command: "kubectl get service web-svc"},
	}
}
