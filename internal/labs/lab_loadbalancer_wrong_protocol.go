package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&LoadBalancerWrongProtocol{})
}

type LoadBalancerWrongProtocol struct {
	BaseLab
}

func (l *LoadBalancerWrongProtocol) ID() string             { return "loadbalancer_wrong_protocol" }
func (l *LoadBalancerWrongProtocol) Title() string          { return "LoadBalancer Wrong Protocol" }
func (l *LoadBalancerWrongProtocol) Category() Category     { return CategoryNetworking }
func (l *LoadBalancerWrongProtocol) Difficulty() Difficulty { return DifficultyMedium }
func (l *LoadBalancerWrongProtocol) EstimatedTime() int     { return 15 }
func (l *LoadBalancerWrongProtocol) Tags() []string {
	return []string{"networking", "loadbalancer", "protocol"}
}

func (l *LoadBalancerWrongProtocol) Description() string {
	return `A LoadBalancer service is using the wrong protocol for health checks.
Traffic is being dropped. Fix the protocol configuration.`
}

func (l *LoadBalancerWrongProtocol) Hints() []string {
	return []string{
		"Check the service annotations",
		"Look at the protocol field in the port spec",
		"Verify the backend application protocol",
	}
}

func (l *LoadBalancerWrongProtocol) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *LoadBalancerWrongProtocol) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Service
metadata:
  name: web-lb
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-health-check-protocol: HTTP
    service.beta.kubernetes.io/aws-load-balancer-health-check-port: "8080"
spec:
  type: LoadBalancer
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 80
    protocol: UDP
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

func (l *LoadBalancerWrongProtocol) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "web-lb",
		"-o", "jsonpath={.spec.ports[0].protocol}")
	if err != nil {
		return err
	}
	if output == "UDP" {
		return fmt.Errorf("protocol still UDP")
	}
	return nil
}

func (l *LoadBalancerWrongProtocol) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check service", Command: "kubectl get svc web-lb -o yaml"},
		{Description: "Fix protocol", Command: "kubectl patch svc web-lb -p '{\"spec\":{\"ports\":[{\"port\":80,\"targetPort\":80,\"protocol\":\"TCP\"}]}}'"},
	}
}
