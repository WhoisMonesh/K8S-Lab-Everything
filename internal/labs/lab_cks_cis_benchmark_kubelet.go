package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSCISBenchmarkKubeletLab{})
}

type CKSCISBenchmarkKubeletLab struct {
	BaseLab
}

func (l *CKSCISBenchmarkKubeletLab) ID() string             { return "cks_cis_benchmark_kubelet" }
func (l *CKSCISBenchmarkKubeletLab) Title() string          { return "Harden Kubelet with CIS Benchmarks" }
func (l *CKSCISBenchmarkKubeletLab) Category() Category     { return CategoryClusterSetupCKS }
func (l *CKSCISBenchmarkKubeletLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSCISBenchmarkKubeletLab) EstimatedTime() int     { return 30 }
func (l *CKSCISBenchmarkKubeletLab) Cert() Cert             { return CertCKS }
func (l *CKSCISBenchmarkKubeletLab) DomainWeight() int      { return 15 }
func (l *CKSCISBenchmarkKubeletLab) Tags() []string {
	return []string{"cks", "kubelet", "cis-benchmark", "hardening"}
}

func (l *CKSCISBenchmarkKubeletLab) Description() string {
	return `The kubelet configuration on worker nodes does not comply with CIS benchmarks.
Several security settings are missing or misconfigured.

Your task: Ensure the kubelet configuration includes:
- anonymous-auth disabled
- authorization-mode set to Webhook
- clientCaFile configured
- tlsCertFile and tlsPrivateKeyFile configured`
}

func (l *CKSCISBenchmarkKubeletLab) Hints() []string {
	return []string{
		"Check /var/lib/kubelet/config.yaml on worker nodes",
		"Reference CIS Kubernetes Benchmark for kubelet settings",
		"Restart kubelet after changes",
	}
}

func (l *CKSCISBenchmarkKubeletLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSCISBenchmarkKubeletLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("getting control plane node: %w", err)
	}
	_, err = dockerCommand(nodeName, "sed -i 's/anonymous-auth: false/anonymous-auth: true/' /var/lib/kubelet/config.yaml 2>/dev/null; "+
		"sed -i 's/mode: Webhook/mode: AlwaysAllow/' /var/lib/kubelet/config.yaml 2>/dev/null; true")
	return err
}

func (l *CKSCISBenchmarkKubeletLab) Verify(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	output, err := dockerCommand(nodeName, "cat /var/lib/kubelet/config.yaml")
	if err != nil {
		return fmt.Errorf("could not read kubelet config: %w", err)
	}
	if strings.Contains(output, "anonymous-auth: true") {
		return fmt.Errorf("anonymous-auth is still enabled")
	}
	if strings.Contains(output, "mode: AlwaysAllow") {
		return fmt.Errorf("authorization mode is still AlwaysAllow")
	}
	return nil
}

func (l *CKSCISBenchmarkKubeletLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current kubelet config", Command: "cat /var/lib/kubelet/config.yaml"},
		{Description: "Edit kubelet config for CIS compliance", Command: "sudo vi /var/lib/kubelet/config.yaml"},
		{Description: "Add required settings", Command: "anonymous:\n  enabled: false\nauthorization:\n  mode: Webhook\nauthentication:\n  x509:\n    clientCAFile: /etc/kubernetes/pki/ca.crt\ntlsCertFile: /etc/kubernetes/pki/kubelet.crt\ntlsPrivateKeyFile: /etc/kubernetes/pki/kubelet.key"},
		{Description: "Restart kubelet", Command: "sudo systemctl daemon-reload && sudo systemctl restart kubelet"},
	}
}
