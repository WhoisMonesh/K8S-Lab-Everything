package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ExternalDNSNotWorking{})
}

type ExternalDNSNotWorking struct {
	BaseLab
}

func (l *ExternalDNSNotWorking) ID() string            { return "external_dns_not_working" }
func (l *ExternalDNSNotWorking) Title() string         { return "External DNS Resolution Failing" }
func (l *ExternalDNSNotWorking) Category() Category    { return CategoryDNS }
func (l *ExternalDNSNotWorking) Difficulty() Difficulty { return DifficultyMedium }
func (l *ExternalDNSNotWorking) EstimatedTime() int    { return 20 }
func (l *ExternalDNSNotWorking) Tags() []string        { return []string{"dns", "coredns", "networking"} }

func (l *ExternalDNSNotWorking) Description() string {
	return `Pods cannot resolve external DNS names like google.com.
The CoreDNS configuration has been modified incorrectly. Fix the CoreDNS ConfigMap.`
}

func (l *ExternalDNSNotWorking) Hints() []string {
	return []string{
		"Check the CoreDNS ConfigMap in kube-system",
		"Look for the forward plugin configuration",
		"Verify with a DNS lookup test pod",
	}
}

func (l *ExternalDNSNotWorking) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ExternalDNSNotWorking) Break(ctx context.Context, kubeconfigPath string) error {
	cm := `apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health {
           lameduck 5s
        }
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }`
	return kubectlApply(ctx, kubeconfigPath, cm)
}

func (l *ExternalDNSNotWorking) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "run", "dns-test", "--image=busybox:1.36",
		"--rm", "-it", "--restart=Never", "--", "nslookup", "google.com")
	if err != nil {
		return fmt.Errorf("DNS test failed: %w", err)
	}
	if output == "" {
		return fmt.Errorf("DNS resolution failed")
	}
	return nil
}

func (l *ExternalDNSNotWorking) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check CoreDNS ConfigMap", Command: "kubectl get configmap coredns -n kube-system -o yaml"},
		{Description: "Fix forward plugin", Command: "kubectl edit configmap coredns -n kube-system"},
		{Description: "Restart CoreDNS", Command: "kubectl rollout restart deployment coredns -n kube-system"},
	}
}
