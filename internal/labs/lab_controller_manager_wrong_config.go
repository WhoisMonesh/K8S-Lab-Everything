package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ControllerManagerWrongConfig{})
}

type ControllerManagerWrongConfig struct {
	BaseLab
}

func (l *ControllerManagerWrongConfig) ID() string            { return "controller_manager_wrong_config" }
func (l *ControllerManagerWrongConfig) Title() string         { return "Controller Manager Misconfiguration" }
func (l *ControllerManagerWrongConfig) Category() Category    { return CategoryControlPlane }
func (l *ControllerManagerWrongConfig) Difficulty() Difficulty { return DifficultyHard }
func (l *ControllerManagerWrongConfig) EstimatedTime() int    { return 25 }
func (l *ControllerManagerWrongConfig) Tags() []string        { return []string{"controller-manager", "cluster"} }

func (l *ControllerManagerWrongConfig) Description() string {
	return `The kube-controller-manager has incorrect cluster CIDR configuration causing pods to fail networking.
Fix the --cluster-cidr flag to use the correct CIDR range.`
}

func (l *ControllerManagerWrongConfig) Hints() []string {
	return []string{
		"Check kube-controller-manager static pod manifest",
		"Look at --cluster-cidr flag value",
		"Verify with kubectl get nodes -o jsonpath",
	}
}

func (l *ControllerManagerWrongConfig) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ControllerManagerWrongConfig) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *ControllerManagerWrongConfig) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system", "-l", "component=kube-controller-manager",
		"-o", "jsonpath={.items[0].spec.containers[0].args}")
	if err != nil {
		return err
	}
	if containsAny(output, "10.245.0.0/16") {
		return fmt.Errorf("cluster-cidr is still misconfigured")
	}
	return nil
}

func (l *ControllerManagerWrongConfig) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check controller-manager config", Command: "kubectl describe pods -n kube-system -l component=kube-controller-manager"},
		{Description: "Verify node pod CIDR", Command: "kubectl get nodes -o jsonpath='{.items[0].spec.podCIDR}'"},
		{Description: "Fix cluster-cidr", Command: "Edit /etc/kubernetes/manifests/kube-controller-manager.yaml and fix --cluster-cidr"},
	}
}
