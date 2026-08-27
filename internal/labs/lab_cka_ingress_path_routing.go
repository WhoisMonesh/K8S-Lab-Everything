package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&IngressPathRoutingLab{})
}

type IngressPathRoutingLab struct {
	BaseLab
}

func (l *IngressPathRoutingLab) ID() string { return "cka_ingress_path_routing" }
func (l *IngressPathRoutingLab) Title() string {
	return "Configure Path-Based Routing"
}
func (l *IngressPathRoutingLab) Category() Category     { return CategoryServicesNetworking }
func (l *IngressPathRoutingLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *IngressPathRoutingLab) EstimatedTime() int     { return 20 }
func (l *IngressPathRoutingLab) Tags() []string {
	return []string{"ingress", "path-based", "routing"}
}
func (l *IngressPathRoutingLab) Cert() Cert        { return CertCKA }
func (l *IngressPathRoutingLab) DomainWeight() int { return 20 }

func (l *IngressPathRoutingLab) Description() string {
	return `Configure an Ingress with path-based routing so that /api routes to
the api-service and /web routes to the web-service, both on port 80.`
}

func (l *IngressPathRoutingLab) Hints() []string {
	return []string{
		"Add multiple paths in the Ingress spec",
		"Use pathType: Prefix for path matching",
		"Set appropriate backend service names",
	}
}

func (l *IngressPathRoutingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *IngressPathRoutingLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *IngressPathRoutingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "path-routing",
		"-n", "routing-ns", "-o", "jsonpath={.spec.rules[0].http.paths}")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "/api") || !strings.Contains(output, "/web") {
		return fmt.Errorf("path-based routing not configured")
	}
	return nil
}

func (l *IngressPathRoutingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check Ingress", Command: "kubectl get ingress path-routing -n routing-ns -o yaml"},
		{Description: "Patch Ingress", Command: "cat <<EOF | kubectl apply -f -\napiVersion: networking.k8s.io/v1\nkind: Ingress\nmetadata:\n  name: path-routing\n  namespace: routing-ns\nspec:\n  rules:\n  - host: app.example.com\n    http:\n      paths:\n      - path: /api\n        pathType: Prefix\n        backend:\n          service:\n            name: api-service\n            port:\n              number: 80\n      - path: /web\n        pathType: Prefix\n        backend:\n          service:\n            name: web-service\n            port:\n              number: 80\nEOF"},
		{Description: "Verify", Command: "kubectl get ingress path-routing -n routing-ns -o yaml"},
	}
}
