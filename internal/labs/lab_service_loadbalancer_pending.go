package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&ServiceLoadBalancerPendingLab{}) }

type ServiceLoadBalancerPendingLab struct{ BaseLab }

func (l *ServiceLoadBalancerPendingLab) ID() string             { return "service_loadbalancer_pending" }
func (l *ServiceLoadBalancerPendingLab) Title() string          { return "LoadBalancer Service Stuck Pending" }
func (l *ServiceLoadBalancerPendingLab) Category() Category     { return CategoryNetworking }
func (l *ServiceLoadBalancerPendingLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *ServiceLoadBalancerPendingLab) EstimatedTime() int     { return 10 }
func (l *ServiceLoadBalancerPendingLab) Tags() []string {
	return []string{"service", "loadbalancer", "metalb", "networking"}
}
func (l *ServiceLoadBalancerPendingLab) Description() string {
	return `A LoadBalancer Service named 'lb-app' is stuck in Pending status
because no cloud provider or MetalLB is installed to assign an
external IP.

Your task: Change the Service type from LoadBalancer to NodePort so
it can receive traffic without an external load balancer.`
}
func (l *ServiceLoadBalancerPendingLab) Hints() []string {
	return []string{
		"Check: kubectl get svc lb-app — TYPE is LoadBalancer, EXTERNAL-IP is <pending>",
		"No MetalLB is installed in this cluster",
		"Patch: kubectl patch svc lb-app -p '{\"spec\":{\"type\":\"NodePort\"}}'",
	}
}

func (l *ServiceLoadBalancerPendingLab) Break(ctx context.Context, kp string) error {
	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: lb-app
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: lb-app
  template:
    metadata:
      labels:
        app: lb-app
    spec:
      containers:
      - name: web
        image: nginx:1.27-alpine
        ports:
        - containerPort: 80
`
	svc := `apiVersion: v1
kind: Service
metadata:
  name: lb-app
  namespace: default
spec:
  type: LoadBalancer
  selector:
    app: lb-app
  ports:
  - port: 80
    targetPort: 80
`
	kubectlApply(ctx, kp, deploy)
	return kubectlApply(ctx, kp, svc)
}

func (l *ServiceLoadBalancerPendingLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *ServiceLoadBalancerPendingLab) Verify(ctx context.Context, kp string) error {
	svcType, _ := kubectl(ctx, kp, "get", "svc", "lb-app", "-o", "jsonpath={.spec.type}")
	if svcType == "LoadBalancer" {
		return fmt.Errorf("service is still type LoadBalancer")
	}
	ready, _ := kubectl(ctx, kp, "get", "deploy", "lb-app", "-o", "jsonpath={.status.readyReplicas}")
	if ready != "2" {
		return fmt.Errorf("deployment not ready (ready: %s)", ready)
	}
	return nil
}

func (l *ServiceLoadBalancerPendingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check the pending Service", Command: "kubectl get svc lb-app", Notes: "TYPE=LoadBalancer, EXTERNAL-IP=<pending>"},
		{Description: "Change to NodePort", Command: `kubectl patch svc lb-app -p '{"spec":{"type":"NodePort"}}'`, Notes: "Instantly assigns a NodePort"},
		{Description: "Verify", Command: "kubectl get svc lb-app", Notes: "TYPE=NodePort, ports show nodePort range"},
	}
}
