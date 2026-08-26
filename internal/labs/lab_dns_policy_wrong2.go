package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&DNSPolicyWrongLab2{})
}

type DNSPolicyWrongLab2 struct {
	BaseLab
}

func (l *DNSPolicyWrongLab2) ID() string             { return "dns_policy_wrong2" }
func (l *DNSPolicyWrongLab2) Title() string          { return "Pod DNS Resolution Broken" }
func (l *DNSPolicyWrongLab2) Category() Category     { return CategoryDNS }
func (l *DNSPolicyWrongLab2) Difficulty() Difficulty { return DifficultyMedium }
func (l *DNSPolicyWrongLab2) EstimatedTime() int     { return 15 }
func (l *DNSPolicyWrongLab2) Tags() []string         { return []string{"dns", "policy", "networking"} }

func (l *DNSPolicyWrongLab2) Description() string {
	return `A pod cannot resolve any DNS names because its DNS policy is set incorrectly.
Fix the DNS policy to use cluster DNS.`
}

func (l *DNSPolicyWrongLab2) Hints() []string {
	return []string{
		"Check the pod DNS policy",
		"Verify the dnsPolicy field",
		"Set dnsPolicy to ClusterFirst",
	}
}

func (l *DNSPolicyWrongLab2) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DNSPolicyWrongLab2) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Pod
metadata:
  name: dns-test
spec:
  dnsPolicy: Default
  containers:
  - name: busybox
    image: busybox:1.36
    command: ["sleep", "3600"]`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *DNSPolicyWrongLab2) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "dns-test",
		"-o", "jsonpath={.spec.dnsPolicy}")
	if err != nil {
		return err
	}
	if output == "Default" {
		return fmt.Errorf("DNS policy still set to Default")
	}
	return nil
}

func (l *DNSPolicyWrongLab2) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod DNS policy", Command: "kubectl get pod dns-test -o jsonpath='{.spec.dnsPolicy}'"},
		{Description: "Fix DNS policy", Command: "kubectl patch pod dns-test -p '{\"spec\":{\"dnsPolicy\":\"ClusterFirst\"}}'"},
		{Description: "Test DNS", Command: "kubectl exec dns-test -- nslookup kubernetes.default.svc.cluster.local"},
	}
}
