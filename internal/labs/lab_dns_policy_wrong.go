package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&DNSPolicyWrongLab{}) }

type DNSPolicyWrongLab struct{ BaseLab }

func (l *DNSPolicyWrongLab) ID() string             { return "dns_policy_wrong" }
func (l *DNSPolicyWrongLab) Title() string          { return "Pod Cannot Resolve Cluster DNS" }
func (l *DNSPolicyWrongLab) Category() Category     { return CategoryDNS }
func (l *DNSPolicyWrongLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *DNSPolicyWrongLab) EstimatedTime() int     { return 15 }
func (l *DNSPolicyWrongLab) Tags() []string {
	return []string{"dns", "dnsPolicy", "networking"}
}
func (l *DNSPolicyWrongLab) Description() string {
	return `A pod 'dns-test' in namespace 'dnsns' cannot resolve any cluster
services. It has dnsPolicy: None with no custom DNS config set, so
it has no nameserver at all.

Your task: Fix the DNS policy to use the cluster DNS (CoreDNS) so the
pod can resolve service names like kubernetes.default.svc.cluster.local.`
}
func (l *DNSPolicyWrongLab) Hints() []string {
	return []string{
		"Check: kubectl get pod dns-test -n dnsns -o yaml | grep dnsPolicy",
		"dnsPolicy: None means 'no DNS' unless you provide dnsConfig",
		"For cluster DNS, use dnsPolicy: ClusterFirst",
		"Delete and recreate the pod with the correct policy",
	}
}

func (l *DNSPolicyWrongLab) Break(ctx context.Context, kp string) error {
	if _, err := kubectl(ctx, kp, "create", "ns", "dnsns"); err != nil {
		return err
	}
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: dns-test
  namespace: dnsns
spec:
  dnsPolicy: None
  containers:
  - name: resolver
    image: busybox:1.36
    command: ["sh","-c","nslookup kubernetes.default.svc.cluster.local 2>&1; sleep 999"]
`
	return kubectlApply(ctx, kp, pod)
}

func (l *DNSPolicyWrongLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *DNSPolicyWrongLab) Verify(ctx context.Context, kp string) error {
	dnsPolicy, _ := kubectl(ctx, kp, "get", "pod", "dns-test", "-n", "dnsns", "-o",
		"jsonpath={.spec.dnsPolicy}")
	if dnsPolicy == "None" {
		return fmt.Errorf("dnsPolicy is still None")
	}
	logs, _ := kubectl(ctx, kp, "logs", "dns-test", "-n", "dnsns", "--tail=3")
	if strings.Contains(logs, "Address 1") || strings.Contains(logs, "Name:") {
		return nil // DNS resolving
	}
	return fmt.Errorf("DNS not working yet — logs: %s", logs)
}

func (l *DNSPolicyWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check dnsPolicy", Command: "kubectl get pod dns-test -n dnsns -o jsonpath='{.spec.dnsPolicy}'", Notes: "Shows None"},
		{Description: "Test DNS failure", Command: "kubectl logs dns-test -n dnsns --tail=5", Notes: "nslookup fails — no nameserver configured"},
		{Description: "Delete and recreate with correct policy", Command: "kubectl delete pod dns-test -n dnsns", Notes: "Pod is immutable for dnsPolicy"},
		{Description: "Recreate with ClusterFirst", Command: `kubectl run dns-test -n dnsns --image=busybox:1.36 --dnsPolicy=ClusterFirst -- sh -c "nslookup kubernetes.default.svc.cluster.local; sleep 999"`, Notes: "Or apply a YAML with dnsPolicy: ClusterFirst"},
		{Description: "Verify DNS works", Command: "kubectl logs dns-test -n dnsns --tail=5", Notes: "Shows Address 10.96.0.10 — CoreDNS responds"},
	}
}
