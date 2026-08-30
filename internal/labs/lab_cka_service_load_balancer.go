package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&ServiceLoadBalancerLab{})
}

type ServiceLoadBalancerLab struct {
	BaseLab
}

func (l *ServiceLoadBalancerLab) ID() string { return "cka_service_load_balancer" }
func (l *ServiceLoadBalancerLab) Title() string {
	return "Configure LoadBalancer Service"
}
func (l *ServiceLoadBalancerLab) Category() Category     { return CategoryServicesNetworking }
func (l *ServiceLoadBalancerLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *ServiceLoadBalancerLab) EstimatedTime() int     { return 20 }
func (l *ServiceLoadBalancerLab) Tags() []string {
	return []string{"service", "loadbalancer", "networking"}
}
func (l *ServiceLoadBalancerLab) Cert() Cert        { return CertCKA }
func (l *ServiceLoadBalancerLab) DomainWeight() int { return 20 }

func (l *ServiceLoadBalancerLab) Description() string {
	return `A LoadBalancer service is not properly configured with the correct
ports and target ports. Fix the service definition to properly route
traffic to the backend pods.`
}

func (l *ServiceLoadBalancerLab) Hints() []string {
	return []string{
		"Check the service spec ports",
		"Verify targetPort matches container port",
		"Ensure protocol is correct",
	}
}

func (l *ServiceLoadBalancerLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceLoadBalancerLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: lb-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: lb-ns
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
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 8080
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	svc := `apiVersion: v1
kind: Service
metadata:
  name: web-lb
  namespace: lb-ns
spec:
  type: LoadBalancer
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 9999
`
	if err := kubectlApply(ctx, kubeconfigPath, svc); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}

func (l *ServiceLoadBalancerLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ServiceLoadBalancerLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "web-lb",
		"-n", "lb-ns", "-o", "jsonpath={.spec.ports[0].targetPort}")
	if err != nil {
		return err
	}
	if output == "80" || output == "8080" {
		return nil
	}
	return fmt.Errorf("targetPort not configured correctly")
}

func (l *ServiceLoadBalancerLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check service", Command: "kubectl get svc web-lb -n lb-ns -o yaml"},
		{Description: "Fix ports", Command: "kubectl patch svc web-lb -n lb-ns -p '{\"spec\":{\"ports\":[{\"port\":80,\"targetPort\":8080,\"protocol\":\"TCP\"}]}}'"},
		{Description: "Verify", Command: "kubectl get svc web-lb -n lb-ns"},
	}
}
