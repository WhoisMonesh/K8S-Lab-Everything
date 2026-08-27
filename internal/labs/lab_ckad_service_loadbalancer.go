package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADServiceLoadBalancerLab{})
}

type CKADServiceLoadBalancerLab struct {
	BaseLab
}

func (l *CKADServiceLoadBalancerLab) ID() string {
	return "ckad_service_loadbalancer"
}

func (l *CKADServiceLoadBalancerLab) Title() string          { return "Create LoadBalancer Service" }
func (l *CKADServiceLoadBalancerLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADServiceLoadBalancerLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADServiceLoadBalancerLab) Cert() Cert             { return CertCKAD }
func (l *CKADServiceLoadBalancerLab) DomainWeight() int      { return 20 }
func (l *CKADServiceLoadBalancerLab) EstimatedTime() int     { return 10 }
func (l *CKADServiceLoadBalancerLab) Tags() []string {
	return []string{"service", "loadbalancer", "external"}
}

func (l *CKADServiceLoadBalancerLab) Description() string {
	return `An application needs to be exposed externally using a LoadBalancer service.
Create a LoadBalancer service named 'public-svc'.

Your task: Create the LoadBalancer service for the web deployment.`
}

func (l *CKADServiceLoadBalancerLab) Hints() []string {
	return []string{
		"LoadBalancer provisions an external load balancer",
		"Use --type=LoadBalancer",
		"Cloud providers assign an external IP",
	}
}

func (l *CKADServiceLoadBalancerLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADServiceLoadBalancerLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: public-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: public-app
  template:
    metadata:
      labels:
        app: public-app
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *CKADServiceLoadBalancerLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "public-svc",
		"-o", "jsonpath={.spec.type}")
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}
	if strings.TrimSpace(output) != "LoadBalancer" {
		return fmt.Errorf("service type is not LoadBalancer (current: %s)", output)
	}
	return nil
}

func (l *CKADServiceLoadBalancerLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create service", Command: "kubectl expose deployment public-app --name=public-svc --port=80 --target-port=80 --type=LoadBalancer"},
		{Description: "Verify", Command: "kubectl get service public-svc"},
	}
}
