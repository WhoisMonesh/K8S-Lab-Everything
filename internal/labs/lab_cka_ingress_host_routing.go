package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&IngressHostRoutingLab{})
}

type IngressHostRoutingLab struct {
	BaseLab
}

func (l *IngressHostRoutingLab) ID() string { return "cka_ingress_host_routing" }
func (l *IngressHostRoutingLab) Title() string {
	return "Configure Host-Based Routing"
}
func (l *IngressHostRoutingLab) Category() Category     { return CategoryServicesNetworking }
func (l *IngressHostRoutingLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *IngressHostRoutingLab) EstimatedTime() int     { return 20 }
func (l *IngressHostRoutingLab) Tags() []string {
	return []string{"ingress", "host-based", "routing"}
}
func (l *IngressHostRoutingLab) Cert() Cert        { return CertCKA }
func (l *IngressHostRoutingLab) DomainWeight() int { return 20 }

func (l *IngressHostRoutingLab) Description() string {
	return `Configure an Ingress with host-based routing so that api.example.com
routes to the api-service and web.example.com routes to the web-service.`
}

func (l *IngressHostRoutingLab) Hints() []string {
	return []string{
		"Add multiple host rules in the Ingress spec",
		"Each host should route to a different backend",
		"Ensure the Ingress controller supports host-based routing",
	}
}

func (l *IngressHostRoutingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *IngressHostRoutingLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: routing-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	svc := `apiVersion: v1
kind: Service
metadata:
  name: api-service
  namespace: routing-ns
spec:
  selector:
    app: api
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, svc); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	ingress := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: host-routing
  namespace: routing-ns
spec:
  rules:
  - host: wrong.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-service
            port:
              number: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, ingress); err != nil {
		return fmt.Errorf("creating ingress: %w", err)
	}

	return nil
}

func (l *IngressHostRoutingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *IngressHostRoutingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "host-routing",
		"-n", "routing-ns", "-o", "jsonpath={.spec.rules}")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "api.example.com") || !strings.Contains(output, "web.example.com") {
		return fmt.Errorf("host-based routing not configured")
	}
	return nil
}

func (l *IngressHostRoutingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check Ingress", Command: "kubectl get ingress host-routing -n routing-ns -o yaml"},
		{Description: "Patch Ingress", Command: "cat <<EOF | kubectl apply -f -\napiVersion: networking.k8s.io/v1\nkind: Ingress\nmetadata:\n  name: host-routing\n  namespace: routing-ns\nspec:\n  rules:\n  - host: api.example.com\n    http:\n      paths:\n      - path: /\n        pathType: Prefix\n        backend:\n          service:\n            name: api-service\n            port:\n              number: 80\n  - host: web.example.com\n    http:\n      paths:\n      - path: /\n        pathType: Prefix\n        backend:\n          service:\n            name: web-service\n            port:\n              number: 80\nEOF"},
		{Description: "Verify", Command: "kubectl get ingress host-routing -n routing-ns -o yaml"},
	}
}
