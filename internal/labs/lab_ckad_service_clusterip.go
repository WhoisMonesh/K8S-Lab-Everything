package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADServiceClusterIPLab{})
}

type CKADServiceClusterIPLab struct {
	BaseLab
}

func (l *CKADServiceClusterIPLab) ID() string             { return "ckad_service_clusterip" }
func (l *CKADServiceClusterIPLab) Title() string          { return "Create ClusterIP Service" }
func (l *CKADServiceClusterIPLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADServiceClusterIPLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADServiceClusterIPLab) Cert() Cert             { return CertCKAD }
func (l *CKADServiceClusterIPLab) DomainWeight() int      { return 20 }
func (l *CKADServiceClusterIPLab) EstimatedTime() int     { return 10 }
func (l *CKADServiceClusterIPLab) Tags() []string {
	return []string{"service", "clusterip", "internal"}
}

func (l *CKADServiceClusterIPLab) Description() string {
	return `A backend application needs to be accessible within the cluster. Create
a ClusterIP service named 'backend-svc' for the backend deployment.

Your task: Create the ClusterIP service exposing port 80.`
}

func (l *CKADServiceClusterIPLab) Hints() []string {
	return []string{
		"ClusterIP is the default service type",
		"Use --type=ClusterIP or omit the type",
		"Match the selector to the deployment labels",
	}
}

func (l *CKADServiceClusterIPLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADServiceClusterIPLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  replicas: 2
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *CKADServiceClusterIPLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "backend-svc",
		"-o", "jsonpath={.spec.type}")
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}
	if strings.TrimSpace(output) != "ClusterIP" && strings.TrimSpace(output) != "" {
		return fmt.Errorf("service type is not ClusterIP (current: %s)", output)
	}
	return nil
}

func (l *CKADServiceClusterIPLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create service", Command: "kubectl expose deployment backend --name=backend-svc --port=80 --target-port=80 --type=ClusterIP"},
		{Description: "Verify", Command: "kubectl get service backend-svc"},
	}
}
