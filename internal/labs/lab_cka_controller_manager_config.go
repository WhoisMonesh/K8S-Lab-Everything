package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&ControllerManagerConfigLab{})
}

type ControllerManagerConfigLab struct {
	BaseLab
}

func (l *ControllerManagerConfigLab) ID() string { return "cka_controller_manager_config" }
func (l *ControllerManagerConfigLab) Title() string {
	return "Fix Controller Manager Arguments"
}
func (l *ControllerManagerConfigLab) Category() Category     { return CategoryClusterArchitecture }
func (l *ControllerManagerConfigLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *ControllerManagerConfigLab) EstimatedTime() int     { return 20 }
func (l *ControllerManagerConfigLab) Tags() []string {
	return []string{"controller-manager", "configuration", "cluster"}
}
func (l *ControllerManagerConfigLab) Cert() Cert        { return CertCKA }
func (l *ControllerManagerConfigLab) DomainWeight() int { return 25 }

func (l *ControllerManagerConfigLab) Description() string {
	return `The kube-controller-manager has an incorrect --service-cluster-ip-range
causing services to get IPs from the wrong range. Fix the controller manager
configuration to use the correct service CIDR.`
}

func (l *ControllerManagerConfigLab) Hints() []string {
	return []string{
		"Check the controller manager static pod manifest",
		"Verify the --service-cluster-ip-range flag",
		"Use kubectl get svc to check current service CIDR",
	}
}

func (l *ControllerManagerConfigLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ControllerManagerConfigLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *ControllerManagerConfigLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-controller-manager", "-o",
		"jsonpath={.items[0].spec.containers[0].args}")
	if err != nil {
		return err
	}
	if strings.Contains(output, "10.245.0.0/16") {
		return fmt.Errorf("controller manager still using wrong service CIDR")
	}
	return nil
}

func (l *ControllerManagerConfigLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check controller manager config", Command: "kubectl describe pods -n kube-system -l component=kube-controller-manager"},
		{Description: "Verify correct service CIDR", Command: "kubectl get svc -A -o wide"},
		{Description: "Edit controller manager manifest", Command: "sudo vi /etc/kubernetes/manifests/kube-controller-manager.yaml"},
		{Description: "Fix --service-cluster-ip-range", Command: "Change to --service-cluster-ip-range=10.96.0.0/12"},
	}
}
