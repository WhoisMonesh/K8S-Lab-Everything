package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CoreDNSUpgradeLab{})
}

type CoreDNSUpgradeLab struct {
	BaseLab
}

func (l *CoreDNSUpgradeLab) ID() string             { return "cka_cluster_dns_upgrade" }
func (l *CoreDNSUpgradeLab) Title() string          { return "Upgrade CoreDNS Deployment" }
func (l *CoreDNSUpgradeLab) Category() Category     { return CategoryClusterArchitecture }
func (l *CoreDNSUpgradeLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CoreDNSUpgradeLab) EstimatedTime() int     { return 20 }
func (l *CoreDNSUpgradeLab) Tags() []string {
	return []string{"coredns", "dns", "upgrade"}
}
func (l *CoreDNSUpgradeLab) Cert() Cert        { return CertCKA }
func (l *CoreDNSUpgradeLab) DomainWeight() int { return 25 }

func (l *CoreDNSUpgradeLab) Description() string {
	return `The CoreDNS deployment is running an outdated version. Upgrade it to the
latest stable version while maintaining DNS service availability. Use a
rolling update strategy.`
}

func (l *CoreDNSUpgradeLab) Hints() []string {
	return []string{
		"Check the current CoreDNS version",
		"Update the deployment image",
		"Ensure readiness probes pass before continuing",
	}
}

func (l *CoreDNSUpgradeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CoreDNSUpgradeLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CoreDNSUpgradeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "coredns",
		"-n", "kube-system", "-o", "jsonpath={.spec.template.spec.containers[0].image}")
	if err != nil {
		return err
	}
	if strings.Contains(output, "1.8.") || strings.Contains(output, "1.9.") || strings.Contains(output, "1.10.") {
		return fmt.Errorf("CoreDNS still on old version: %s", output)
	}
	return nil
}

func (l *CoreDNSUpgradeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current version", Command: "kubectl get deployment coredns -n kube-system -o yaml"},
		{Description: "Update image", Command: "kubectl set image deployment/coredns coredns=coredns/coredns:v1.11.1 -n kube-system"},
		{Description: "Watch rollout", Command: "kubectl rollout status deployment/coredns -n kube-system"},
	}
}
