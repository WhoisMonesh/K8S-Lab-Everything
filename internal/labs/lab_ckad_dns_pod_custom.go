package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADDNSPodCustomLab{})
}

type CKADDNSPodCustomLab struct {
	BaseLab
}

func (l *CKADDNSPodCustomLab) ID() string             { return "ckad_dns_pod_custom" }
func (l *CKADDNSPodCustomLab) Title() string          { return "Customize Pod DNS Settings" }
func (l *CKADDNSPodCustomLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADDNSPodCustomLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADDNSPodCustomLab) Cert() Cert             { return CertCKAD }
func (l *CKADDNSPodCustomLab) DomainWeight() int      { return 20 }
func (l *CKADDNSPodCustomLab) EstimatedTime() int     { return 15 }
func (l *CKADDNSPodCustomLab) Tags() []string {
	return []string{"dns", "pod-dns", "configuration"}
}

func (l *CKADDNSPodCustomLab) Description() string {
	return `A pod needs custom DNS configuration. Configure the pod to use specific
DNS servers and search domains.

Your task: Add dnsPolicy and dnsConfig to the pod spec.`
}

func (l *CKADDNSPodCustomLab) Hints() []string {
	return []string{
		"Use dnsPolicy: None for custom DNS",
		"Add dnsConfig with nameservers and searches",
		"Custom DNS requires explicit nameserver IPs",
	}
}

func (l *CKADDNSPodCustomLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADDNSPodCustomLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: custom-dns
  labels:
    app: custom-dns
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'cat /etc/resolv.conf && sleep 3600']`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADDNSPodCustomLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "custom-dns",
		"-o", "jsonpath={.spec.dnsPolicy}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "" || strings.TrimSpace(output) == "ClusterFirst" {
		return fmt.Errorf("custom DNS policy not configured")
	}
	return nil
}

func (l *CKADDNSPodCustomLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod", Command: "kubectl edit pod custom-dns"},
		{Description: "Add DNS config", Command: "Add dnsPolicy: None and dnsConfig with nameservers and searches"},
		{Description: "Verify", Command: "kubectl exec custom-dns -- cat /etc/resolv.conf"},
	}
}
