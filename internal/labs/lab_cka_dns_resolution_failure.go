package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&DNSResolutionFailureLab{})
}

type DNSResolutionFailureLab struct {
	BaseLab
}

func (l *DNSResolutionFailureLab) ID() string { return "cka_dns_resolution_failure" }
func (l *DNSResolutionFailureLab) Title() string {
	return "Debug DNS Resolution Failures"
}
func (l *DNSResolutionFailureLab) Category() Category     { return CategoryTroubleshooting }
func (l *DNSResolutionFailureLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *DNSResolutionFailureLab) EstimatedTime() int     { return 25 }
func (l *DNSResolutionFailureLab) Tags() []string {
	return []string{"dns", "coredns", "resolution", "troubleshooting"}
}
func (l *DNSResolutionFailureLab) Cert() Cert        { return CertCKA }
func (l *DNSResolutionFailureLab) DomainWeight() int { return 30 }

func (l *DNSResolutionFailureLab) Description() string {
	return `Pods cannot resolve service names. Debug DNS resolution by checking
CoreDNS pods, service configuration, and resolv.conf settings in
affected pods.`
}

func (l *DNSResolutionFailureLab) Hints() []string {
	return []string{
		"Check CoreDNS pod status",
		"Verify the kube-dns service exists",
		"Check pod's /etc/resolv.conf",
	}
}

func (l *DNSResolutionFailureLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DNSResolutionFailureLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *DNSResolutionFailureLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "-n", "dns-ns", "dns-test-pod",
		"--", "nslookup", "kubernetes.default.svc.cluster.local")
	if err != nil {
		return fmt.Errorf("DNS resolution still failing: %w", err)
	}
	if strings.Contains(output, "NXDOMAIN") || strings.Contains(output, "SERVFAIL") {
		return fmt.Errorf("DNS resolution failed")
	}
	return nil
}

func (l *DNSResolutionFailureLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check CoreDNS pods", Command: "kubectl get pods -n kube-system -l k8s-app=kube-dns"},
		{Description: "Check kube-dns service", Command: "kubectl get svc kube-dns -n kube-system"},
		{Description: "Test DNS", Command: "kubectl exec -n dns-ns dns-test-pod -- nslookup kubernetes.default"},
		{Description: "Fix CoreDNS", Command: "Restart or fix CoreDNS deployment"},
	}
}
