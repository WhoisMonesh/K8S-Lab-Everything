package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
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
configuration to use the correct service CIDR (10.96.0.0/12).

kind nodes are containers (no SSH); access the control-plane node shell with:
    docker exec -it <cluster>-control-plane bash`
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
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	output, err := dockerExec(ctx, nodeName, "cat", "/etc/kubernetes/manifests/kube-controller-manager.yaml")
	if err != nil {
		return fmt.Errorf("reading kube-controller-manager manifest: %w", err)
	}

	modifiedManifest := strings.Replace(output,
		"- kube-controller-manager",
		"- kube-controller-manager\n    - --service-cluster-ip-range=10.245.0.0/16",
		1)

	writeCmd := fmt.Sprintf("cat > /etc/kubernetes/manifests/kube-controller-manager.yaml << 'EOF'\n%s\nEOF", modifiedManifest)
	_, err = dockerExec(ctx, nodeName, "sh", "-c", writeCmd)
	if err != nil {
		return fmt.Errorf("writing modified manifest: %w", err)
	}

	return nil
}

func (l *ControllerManagerConfigLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
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
		{Description: "Check controller manager pod status", Command: "kubectl get pods -n kube-system | grep controller-manager"},
		{Description: "Access the control plane node", Command: "docker exec -it <cluster>-control-plane bash"},
		{Description: "Examine the controller manager manifest", Command: "cat /etc/kubernetes/manifests/kube-controller-manager.yaml"},
		{Description: "Remove or fix the wrong service CIDR flag", Command: "sed -i '/--service-cluster-ip-range=10.245.0.0\\/16/d' /etc/kubernetes/manifests/kube-controller-manager.yaml"},
		{Description: "Verify controller manager is running with correct CIDR", Command: "kubectl get pods -n kube-system | grep controller-manager"},
	}
}
