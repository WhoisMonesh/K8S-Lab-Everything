package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ServiceLoadBalancerPending2{})
}

type ServiceLoadBalancerPending2 struct {
	BaseLab
}

func (l *ServiceLoadBalancerPending2) ID() string             { return "service_loadbalancer_pending2" }
func (l *ServiceLoadBalancerPending2) Title() string          { return "LoadBalancer Service Stuck Pending" }
func (l *ServiceLoadBalancerPending2) Category() Category     { return CategoryNetworking }
func (l *ServiceLoadBalancerPending2) Difficulty() Difficulty { return DifficultyEasy }
func (l *ServiceLoadBalancerPending2) EstimatedTime() int     { return 10 }
func (l *ServiceLoadBalancerPending2) Tags() []string {
	return []string{"networking", "loadbalancer", "service"}
}

func (l *ServiceLoadBalancerPending2) Description() string {
	return `A LoadBalancer service is stuck in Pending state.
The cloud provider is not provisioning an external load balancer.
Change the service type to NodePort as a workaround.`
}

func (l *ServiceLoadBalancerPending2) Hints() []string {
	return []string{
		"Check the service status",
		"Look at service events",
		"Consider changing to NodePort type",
	}
}

func (l *ServiceLoadBalancerPending2) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceLoadBalancerPending2) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Service
metadata:
  name: lb-pending
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

func (l *ServiceLoadBalancerPending2) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "lb-pending",
		"-o", "jsonpath={.spec.type}")
	if err != nil {
		return err
	}
	if output == "LoadBalancer" {
		return fmt.Errorf("service still LoadBalancer type")
	}
	return nil
}

func (l *ServiceLoadBalancerPending2) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check service", Command: "kubectl get svc lb-pending"},
		{Description: "Check events", Command: "kubectl get events --field-selector involvedObject.name=lb-pending"},
		{Description: "Change to NodePort", Command: "kubectl patch svc lb-pending -p '{\"spec\":{\"type\":\"NodePort\"}}'"},
	}
}
