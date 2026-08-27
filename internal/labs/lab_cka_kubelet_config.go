package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&KubeletConfigLab{})
}

type KubeletConfigLab struct {
	BaseLab
}

func (l *KubeletConfigLab) ID() string             { return "cka_kubelet_config" }
func (l *KubeletConfigLab) Title() string          { return "Fix Kubelet Configuration Issues" }
func (l *KubeletConfigLab) Category() Category     { return CategoryClusterArchitecture }
func (l *KubeletConfigLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *KubeletConfigLab) EstimatedTime() int     { return 20 }
func (l *KubeletConfigLab) Tags() []string {
	return []string{"kubelet", "configuration", "node"}
}
func (l *KubeletConfigLab) Cert() Cert        { return CertCKA }
func (l *KubeletConfigLab) DomainWeight() int { return 25 }

func (l *KubeletConfigLab) Description() string {
	return `A worker node's kubelet is misconfigured with an incorrect --cluster-dns
setting causing DNS resolution failures in pods. Fix the kubelet configuration
to use the correct cluster DNS IP.`
}

func (l *KubeletConfigLab) Hints() []string {
	return []string{
		"Check the kubelet config file at /var/lib/kubelet/config.yaml",
		"Verify the clusterDNS setting",
		"Restart kubelet after fixing",
	}
}

func (l *KubeletConfigLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeletConfigLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *KubeletConfigLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap", "kubelet-config",
		"-n", "kube-system", "-o", "jsonpath={.data.clusterDNS}")
	if err != nil {
		return err
	}
	if strings.Contains(output, "10.96.0.10") {
		return fmt.Errorf("kubelet still using wrong cluster DNS")
	}
	return nil
}

func (l *KubeletConfigLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check kubelet config", Command: "kubectl get configmap kubelet-config -n kube-system -o yaml"},
		{Description: "Get correct CoreDNS IP", Command: "kubectl get svc -n kube-system kube-dns -o jsonpath='{.spec.clusterIP}'"},
		{Description: "Edit kubelet config", Command: "sudo vi /var/lib/kubelet/config.yaml"},
		{Description: "Restart kubelet", Command: "sudo systemctl daemon-reload && sudo systemctl restart kubelet"},
	}
}
