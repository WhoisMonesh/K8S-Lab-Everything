package labs

import (
	"context"
	"fmt"
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
