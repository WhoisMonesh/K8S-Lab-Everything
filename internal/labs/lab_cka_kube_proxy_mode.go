package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&KubeProxyModeLab{})
}

type KubeProxyModeLab struct {
	BaseLab
}

func (l *KubeProxyModeLab) ID() string             { return "cka_kube_proxy_mode" }
func (l *KubeProxyModeLab) Title() string          { return "Switch kube-proxy Between iptables and ipvs" }
func (l *KubeProxyModeLab) Category() Category     { return CategoryClusterArchitecture }
func (l *KubeProxyModeLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *KubeProxyModeLab) EstimatedTime() int     { return 20 }
func (l *KubeProxyModeLab) Tags() []string {
	return []string{"kube-proxy", "ipvs", "iptables", "networking"}
}
func (l *KubeProxyModeLab) Cert() Cert        { return CertCKA }
func (l *KubeProxyModeLab) DomainWeight() int { return 25 }

func (l *KubeProxyModeLab) Description() string {
	return `The kube-proxy is running in iptables mode. Switch it to IPVS mode
for better performance with large numbers of services. Update the ConfigMap
and restart kube-proxy pods.`
}

func (l *KubeProxyModeLab) Hints() []string {
	return []string{
		"Edit the kube-proxy ConfigMap",
		"Change the mode field to IPVS",
		"Delete kube-proxy pods to restart them",
	}
}

func (l *KubeProxyModeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeProxyModeLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *KubeProxyModeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap", "kube-proxy",
		"-n", "kube-system", "-o", "jsonpath={.data.config.conf}")
	if err != nil {
		return err
	}
	if strings.Contains(output, "iptables") {
		return fmt.Errorf("kube-proxy still in iptables mode")
	}
	return nil
}

func (l *KubeProxyModeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current mode", Command: "kubectl get configmap kube-proxy -n kube-system -o yaml"},
		{Description: "Edit ConfigMap", Command: "kubectl edit configmap kube-proxy -n kube-system"},
		{Description: "Change mode to IPVS", Command: "Set mode: \"ipvs\" in the config data"},
		{Description: "Restart kube-proxy", Command: "kubectl delete pods -n kube-system -l k8s-app=kube-proxy"},
	}
}
