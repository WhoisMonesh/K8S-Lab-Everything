package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&KubeletPodMaxPodsLab{})
}

type KubeletPodMaxPodsLab struct {
	BaseLab
}

func (l *KubeletPodMaxPodsLab) ID() string { return "cka_kubelet_pod_max_pods" }
func (l *KubeletPodMaxPodsLab) Title() string {
	return "Configure kubelet maxPods Setting"
}
func (l *KubeletPodMaxPodsLab) Category() Category     { return CategoryClusterArchitecture }
func (l *KubeletPodMaxPodsLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *KubeletPodMaxPodsLab) EstimatedTime() int     { return 15 }
func (l *KubeletPodMaxPodsLab) Tags() []string {
	return []string{"kubelet", "maxpods", "configuration"}
}
func (l *KubeletPodMaxPodsLab) Cert() Cert        { return CertCKA }
func (l *KubeletPodMaxPodsLab) DomainWeight() int { return 25 }

func (l *KubeletPodMaxPodsLab) Description() string {
	return `The kubelet on a worker node has maxPods set too low (10), preventing
additional pods from being scheduled. Increase maxPods to 110 and restart
the kubelet.`
}

func (l *KubeletPodMaxPodsLab) Hints() []string {
	return []string{
		"Edit the kubelet configuration file",
		"Look for maxPods setting",
		"Restart kubelet after changes",
	}
}

func (l *KubeletPodMaxPodsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeletPodMaxPodsLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *KubeletPodMaxPodsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap", "kubelet-config",
		"-n", "kube-system", "-o", "jsonpath={.data.maxPods}")
	if err != nil {
		return err
	}
	if output == "10" {
		return fmt.Errorf("maxPods still set to 10")
	}
	return nil
}

func (l *KubeletPodMaxPodsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current maxPods", Command: "kubectl get configmap kubelet-config -n kube-system -o yaml"},
		{Description: "Edit kubelet config", Command: "kubectl edit configmap kubelet-config -n kube-system"},
		{Description: "Update maxPods", Command: "Change maxPods: 110"},
		{Description: "Restart kubelet", Command: "sudo systemctl daemon-reload && sudo systemctl restart kubelet"},
	}
}
