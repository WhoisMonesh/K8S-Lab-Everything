package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&DNSTimeoutIssues{})
}

type DNSTimeoutIssues struct {
	BaseLab
}

func (l *DNSTimeoutIssues) ID() string             { return "dns_timeout_issues" }
func (l *DNSTimeoutIssues) Title() string          { return "DNS Resolution Timeout" }
func (l *DNSTimeoutIssues) Category() Category     { return CategoryDNS }
func (l *DNSTimeoutIssues) Difficulty() Difficulty { return DifficultyMedium }
func (l *DNSTimeoutIssues) EstimatedTime() int     { return 20 }
func (l *DNSTimeoutIssues) Tags() []string         { return []string{"dns", "timeout", "coredns"} }

func (l *DNSTimeoutIssues) Description() string {
	return `DNS resolution is timing out for services. CoreDNS is overloaded due to misconfigured cache.
Fix the CoreDNS configuration to resolve timeout issues.`
}

func (l *DNSTimeoutIssues) Hints() []string {
	return []string{
		"Check CoreDNS ConfigMap",
		"Look at the cache and forward plugins",
		"Verify CoreDNS pod health",
	}
}

func (l *DNSTimeoutIssues) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DNSTimeoutIssues) Break(ctx context.Context, kubeconfigPath string) error {
	cm := `apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        cache 0
        forward . 192.168.1.1 {
            max_concurrent 1
        }
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
        }
    }`
	return kubectlApply(ctx, kubeconfigPath, cm)
}

func (l *DNSTimeoutIssues) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap", "coredns", "-n", "kube-system",
		"-o", "jsonpath={.data.Corefile}")
	if err != nil {
		return err
	}
	if containsAny(output, "cache 0") {
		return fmt.Errorf("cache still disabled")
	}
	return nil
}

func (l *DNSTimeoutIssues) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check CoreDNS config", Command: "kubectl get configmap coredns -n kube-system -o yaml"},
		{Description: "Fix cache and forward", Command: "kubectl edit configmap coredns -n kube-system"},
		{Description: "Set cache to 30 and increase max_concurrent", Command: "Change cache 0 to cache 30 and max_concurrent to 1000"},
		{Description: "Restart CoreDNS", Command: "kubectl rollout restart deployment coredns -n kube-system"},
	}
}
