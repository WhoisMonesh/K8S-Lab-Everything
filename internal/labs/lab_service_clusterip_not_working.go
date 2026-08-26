package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ServiceClusterIPNotWorking{})
}

type ServiceClusterIPNotWorking struct {
	BaseLab
}

func (l *ServiceClusterIPNotWorking) ID() string            { return "service_clusterip_not_working" }
func (l *ServiceClusterIPNotWorking) Title() string         { return "ClusterIP Service Not Responding" }
func (l *ServiceClusterIPNotWorking) Category() Category    { return CategoryNetworking }
func (l *ServiceClusterIPNotWorking) Difficulty() Difficulty { return DifficultyMedium }
func (l *ServiceClusterIPNotWorking) EstimatedTime() int    { return 15 }
func (l *ServiceClusterIPNotWorking) Tags() []string        { return []string{"service", "clusterip", "networking"} }

func (l *ServiceClusterIPNotWorking) Description() string {
	return `A ClusterIP service is not responding to requests. The service exists but has no endpoints.
Debug and fix the service selector to match the backend pods.`
}

func (l *ServiceClusterIPNotWorking) Hints() []string {
	return []string{
		"Check the service endpoints",
		"Verify the service selector matches pod labels",
		"Check if pods have the correct labels",
	}
}

func (l *ServiceClusterIPNotWorking) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceClusterIPNotWorking) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
        tier: api
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: backend-service
spec:
  selector:
    app: backend
    tier: web
  ports:
  - port: 80
    targetPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ServiceClusterIPNotWorking) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "endpoints", "backend-service",
		"-o", "jsonpath={.subsets[*].addresses[*].ip}")
	if err != nil {
		return err
	}
	if output == "" {
		return fmt.Errorf("service has no endpoints")
	}
	return nil
}

func (l *ServiceClusterIPNotWorking) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check service endpoints", Command: "kubectl get endpoints backend-service"},
		{Description: "Check pod labels", Command: "kubectl get pods --show-labels"},
		{Description: "Fix service selector", Command: "kubectl patch svc backend-service -p '{\"spec\":{\"selector\":{\"tier\":\"api\"}}}'"},
	}
}
