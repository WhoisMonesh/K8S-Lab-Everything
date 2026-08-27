package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADIngressNameBasedLab{})
}

type CKADIngressNameBasedLab struct {
	BaseLab
}

func (l *CKADIngressNameBasedLab) ID() string {
	return "ckad_ingress_name_based"
}

func (l *CKADIngressNameBasedLab) Title() string {
	return "Configure Name-based Virtual Hosting"
}

func (l *CKADIngressNameBasedLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADIngressNameBasedLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADIngressNameBasedLab) Cert() Cert             { return CertCKAD }
func (l *CKADIngressNameBasedLab) DomainWeight() int      { return 20 }
func (l *CKADIngressNameBasedLab) EstimatedTime() int     { return 20 }
func (l *CKADIngressNameBasedLab) Tags() []string {
	return []string{"ingress", "name-based", "virtual-hosting"}
}

func (l *CKADIngressNameBasedLab) Description() string {
	return `An Ingress needs to route traffic based on hostname. Requests to
api.example.com go to the api service and requests to web.example.com
go to the web service.

Your task: Configure name-based virtual hosting in the Ingress resource.`
}

func (l *CKADIngressNameBasedLab) Hints() []string {
	return []string{
		"Define multiple rules with different hosts",
		"Each host routes to a different backend service",
		"Use separate rules for each hostname",
	}
}

func (l *CKADIngressNameBasedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADIngressNameBasedLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADIngressNameBasedLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "vhost-ingress",
		"-o", "jsonpath={.spec.rules[*].host}")
	if err != nil {
		return fmt.Errorf("failed to get ingress: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no hosts found in Ingress")
	}
	if !strings.Contains(output, "api.example.com") || !strings.Contains(output, "web.example.com") {
		return fmt.Errorf("virtual hosting not configured")
	}
	return nil
}

func (l *CKADIngressNameBasedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check services", Command: "kubectl get services"},
		{Description: "Edit Ingress", Command: "Add rule for api.example.com pointing to api service and web.example.com pointing to web service"},
		{Description: "Verify", Command: "kubectl get ingress vhost-ingress -o yaml | grep host"},
	}
}
