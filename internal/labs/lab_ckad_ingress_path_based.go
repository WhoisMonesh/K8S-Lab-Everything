package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADIngressPathBasedLab{})
}

type CKADIngressPathBasedLab struct {
	BaseLab
}

func (l *CKADIngressPathBasedLab) ID() string {
	return "ckad_ingress_path_based"
}

func (l *CKADIngressPathBasedLab) Title() string {
	return "Configure Path-based Routing"
}

func (l *CKADIngressPathBasedLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADIngressPathBasedLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADIngressPathBasedLab) Cert() Cert             { return CertCKAD }
func (l *CKADIngressPathBasedLab) DomainWeight() int      { return 20 }
func (l *CKADIngressPathBasedLab) EstimatedTime() int     { return 20 }
func (l *CKADIngressPathBasedLab) Tags() []string {
	return []string{"ingress", "path-based", "routing"}
}

func (l *CKADIngressPathBasedLab) Description() string {
	return `An Ingress needs to route traffic based on URL paths. Requests to /api
go to the api service and requests to /web go to the web service.

Your task: Configure path-based routing in the Ingress resource.`
}

func (l *CKADIngressPathBasedLab) Hints() []string {
	return []string{
		"Define multiple paths under the same host",
		"Use pathType: Prefix for path matching",
		"Each path points to a different backend service",
	}
}

func (l *CKADIngressPathBasedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADIngressPathBasedLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADIngressPathBasedLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "path-ingress",
		"-o", "jsonpath={.spec.rules[*].http.paths[*].path}")
	if err != nil {
		return fmt.Errorf("failed to get ingress: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no paths found in Ingress")
	}
	if !strings.Contains(output, "/api") || !strings.Contains(output, "/web") {
		return fmt.Errorf("path-based routing not configured")
	}
	return nil
}

func (l *CKADIngressPathBasedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check services", Command: "kubectl get services"},
		{Description: "Edit Ingress", Command: "Add /api path pointing to api service and /web path pointing to web service"},
		{Description: "Verify", Command: "kubectl get ingress path-ingress -o yaml | grep -A 5 paths"},
	}
}
