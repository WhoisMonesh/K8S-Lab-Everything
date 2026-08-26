package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ExternalIPNotAssigned{})
}

type ExternalIPNotAssigned struct {
	BaseLab
}

func (l *ExternalIPNotAssigned) ID() string            { return "external_ip_not_assigned" }
func (l *ExternalIPNotAssigned) Title() string         { return "External IP Not Assigned" }
func (l *ExternalIPNotAssigned) Category() Category    { return CategoryNetworking }
func (l *ExternalIPNotAssigned) Difficulty() Difficulty { return DifficultyMedium }
func (l *ExternalIPNotAssigned) EstimatedTime() int    { return 15 }
func (l *ExternalIPNotAssigned) Tags() []string        { return []string{"networking", "external", "loadbalancer"} }

func (l *ExternalIPNotAssigned) Description() string {
	return `A LoadBalancer service has no external IP assigned.
The cloud provider controller may not be running. Debug and fix the service.`
}

func (l *ExternalIPNotAssigned) Hints() []string {
	return []string{
		"Check the service status",
		"Look for events or errors",
		"Consider using NodePort or NodeBalancer instead",
	}
}

func (l *ExternalIPNotAssigned) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ExternalIPNotAssigned) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Service
metadata:
  name: external-svc
spec:
  type: LoadBalancer
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ExternalIPNotAssigned) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "external-svc",
		"-o", "jsonpath={.status.loadBalancer.ingress[*].ip}")
	if err != nil {
		return err
	}
	if output == "" {
		return fmt.Errorf("no external IP assigned")
	}
	return nil
}

func (l *ExternalIPNotAssigned) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check service", Command: "kubectl get svc external-svc"},
		{Description: "Check events", Command: "kubectl get events --field-selector involvedObject.name=external-svc"},
		{Description: "Change to NodePort", Command: "kubectl patch svc external-svc -p '{\"spec\":{\"type\":\"NodePort\"}}'"},
	}
}
