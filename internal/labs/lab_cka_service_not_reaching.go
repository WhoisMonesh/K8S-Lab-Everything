package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ServiceNotReachingLab{})
}

type ServiceNotReachingLab struct {
	BaseLab
}

func (l *ServiceNotReachingLab) ID() string { return "cka_service_not_reaching" }
func (l *ServiceNotReachingLab) Title() string {
	return "Debug Service Connectivity Issues"
}
func (l *ServiceNotReachingLab) Category() Category     { return CategoryTroubleshooting }
func (l *ServiceNotReachingLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *ServiceNotReachingLab) EstimatedTime() int     { return 20 }
func (l *ServiceNotReachingLab) Tags() []string {
	return []string{"service", "connectivity", "debugging", "troubleshooting"}
}
func (l *ServiceNotReachingLab) Cert() Cert        { return CertCKA }
func (l *ServiceNotReachingLab) DomainWeight() int { return 30 }

func (l *ServiceNotReachingLab) Description() string {
	return `A service is not reachable from pods. Debug the service by checking
endpoints, selector labels, and port configuration to find and fix
the connectivity issue.`
}

func (l *ServiceNotReachingLab) Hints() []string {
	return []string{
		"Check service endpoints",
		"Verify selector matches pod labels",
		"Test connectivity with wget or curl",
	}
}

func (l *ServiceNotReachingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceNotReachingLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *ServiceNotReachingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "endpoints", "web-svc",
		"-n", "svc-ns", "-o", "jsonpath={.subsets[0].addresses}")
	if err != nil {
		return err
	}
	if output == "" || output == "null" {
		return fmt.Errorf("service has no endpoints")
	}
	return nil
}

func (l *ServiceNotReachingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check service", Command: "kubectl get svc web-svc -n svc-ns"},
		{Description: "Check endpoints", Command: "kubectl get endpoints web-svc -n svc-ns"},
		{Description: "Verify selector", Command: "kubectl get svc web-svc -n svc-ns -o yaml | grep -A3 selector"},
		{Description: "Fix selector", Command: "Patch service to match pod labels"},
	}
}
