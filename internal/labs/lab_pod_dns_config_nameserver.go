package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodDNSConfigNameserverLab{})
}

type PodDNSConfigNameserverLab struct {
	BaseLab
}

func (l *PodDNSConfigNameserverLab) ID() string {
	return "pod_dns_config_nameserver"
}

func (l *PodDNSConfigNameserverLab) Title() string {
	return "Pod DNS Nameserver Misconfigured"
}

func (l *PodDNSConfigNameserverLab) Category() Category {
	return CategoryNetworking
}

func (l *PodDNSConfigNameserverLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodDNSConfigNameserverLab) Description() string {
	return `A pod 'dns-test' has a custom DNS config pointing to a non-existent
nameserver (10.0.0.99). This causes DNS resolution to fail for all
external and internal names.

Your task: Fix the pod's DNS configuration to use the cluster DNS.`
}

func (l *PodDNSConfigNameserverLab) Hints() []string {
	return []string{
		"Check the pod's DNS config",
		"The nameserver points to a non-existent IP",
		"Remove the custom dnsConfig or use the cluster DNS IP",
	}
}

func (l *PodDNSConfigNameserverLab) EstimatedTime() int {
	return 10
}

func (l *PodDNSConfigNameserverLab) Tags() []string {
	return []string{"dns", "nameserver", "pod", "networking"}
}

func (l *PodDNSConfigNameserverLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodDNSConfigNameserverLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: dns-test
  namespace: default
spec:
  dnsConfig:
    nameservers:
    - 10.0.0.99
    searches:
    - default.svc.cluster.local
    options:
    - name: ndots
      value: "5"
  containers:
  - name: test
    image: busybox:1.36
    command: ['sh', '-c', 'nslookup kubernetes.default.svc.cluster.local && sleep 3600']
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodDNSConfigNameserverLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodDNSConfigNameserverLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "dns-test",
		"-o", "jsonpath={.spec.dnsConfig.nameservers[0]}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) == "10.0.0.99" {
		return fmt.Errorf("nameserver is still 10.0.0.99")
	}

	// Test DNS resolution
	output, err = kubectl(ctx, kubeconfigPath, "exec", "dns-test",
		"--", "nslookup", "kubernetes.default.svc.cluster.local")
	if err != nil {
		return fmt.Errorf("DNS resolution failed: %w", err)
	}

	if !strings.Contains(output, "Address") {
		return fmt.Errorf("DNS resolution not working")
	}

	return nil
}

func (l *PodDNSConfigNameserverLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod DNS config",
			Command:     "kubectl get pod dns-test -o yaml | grep -A 10 dnsConfig",
			Notes:       "nameserver is 10.0.0.99 which doesn't exist",
		},
		{
			Description: "Test DNS (should fail)",
			Command:     "kubectl exec dns-test -- nslookup kubernetes.default.svc.cluster.local",
			Notes:       "DNS resolution should fail",
		},
		{
			Description: "Fix DNS config",
			Command:     "kubectl edit pod dns-test",
			Notes:       "Remove the custom dnsConfig or change nameserver to cluster DNS",
		},
		{
			Description: "Verify DNS works",
			Command:     "kubectl exec dns-test -- nslookup kubernetes.default.svc.cluster.local",
			Notes:       "Should now resolve successfully",
		},
	}
}
