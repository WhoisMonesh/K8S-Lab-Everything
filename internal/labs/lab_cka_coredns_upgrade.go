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
	return `The CoreDNS deployment is pinned to an outdated version. Upgrade it to the
latest stable version while maintaining DNS service availability. Use a
rolling update strategy and verify all replicas are ready.`
}

func (l *CoreDNSUpgradeLab) Hints() []string {
	return []string{
		"Check the current CoreDNS image version",
		"Update the deployment image with kubectl set image",
		"Watch the rollout and ensure readiness probes pass",
	}
}

func (l *CoreDNSUpgradeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CoreDNSUpgradeLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Real scenario: downgrade CoreDNS to an old, known-bad version so the
	// learner must upgrade it back to a current release.
	current, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "coredns",
		"-n", "kube-system", "-o", "jsonpath={.spec.template.spec.containers[0].image}")
	if err != nil {
		return err
	}
	if strings.Contains(current, "1.8.") || strings.Contains(current, "1.9.") || strings.Contains(current, "1.10.") {
		// Already old (e.g. re-run) — no-op.
		return nil
	}
	if _, err := kubectl(ctx, kubeconfigPath, "set", "image", "deployment/coredns",
		"coredns=coredns/coredns:v1.8.4", "-n", "kube-system"); err != nil {
		return fmt.Errorf("downgrading coredns: %w", err)
	}
	if _, err := kubectl(ctx, kubeconfigPath, "rollout", "status", "deployment/coredns",
		"-n", "kube-system", "--timeout=120s"); err != nil {
		return fmt.Errorf("waiting for coredns downgrade rollout: %w", err)
	}
	return nil
}

func (l *CoreDNSUpgradeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "coredns",
		"-n", "kube-system", "-o", "jsonpath={.spec.template.spec.containers[0].image}")
	if err != nil {
		return err
	}
	for _, old := range []string{"1.8.", "1.9.", "1.10."} {
		if strings.Contains(output, old) {
			return fmt.Errorf("CoreDNS still on old/outdated version: %s", output)
		}
	}
	// Ensure it actually rolled out and is ready.
	if _, err := kubectl(ctx, kubeconfigPath, "rollout", "status", "deployment/coredns",
		"-n", "kube-system", "--timeout=120s"); err != nil {
		return fmt.Errorf("coredns rollout is not complete: %w", err)
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
