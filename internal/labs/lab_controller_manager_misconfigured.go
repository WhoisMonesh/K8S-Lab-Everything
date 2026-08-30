package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ControllerManagerMisconfiguredLab{})
}

type ControllerManagerMisconfiguredLab struct {
	BaseLab
}

func (l *ControllerManagerMisconfiguredLab) ID() string {
	return "controller_manager_misconfigured"
}

func (l *ControllerManagerMisconfiguredLab) Title() string {
	return "Controller Manager Misconfigured"
}

func (l *ControllerManagerMisconfiguredLab) Category() Category {
	return CategoryClusterArchitecture
}

func (l *ControllerManagerMisconfiguredLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *ControllerManagerMisconfiguredLab) Description() string {
	return `The kube-controller-manager has been restarted with an invalid
configuration flag. New deployments and replicasets are not being
reconciled - pods are not being created.

Your task: Diagnose and fix the controller manager configuration to
restore normal cluster operations.`
}

func (l *ControllerManagerMisconfiguredLab) Hints() []string {
	return []string{
		"Check controller manager pod status in kube-system",
		"Look at the controller manager static pod manifest",
		"Check /etc/kubernetes/manifests/kube-controller-manager.yaml",
	}
}

func (l *ControllerManagerMisconfiguredLab) EstimatedTime() int {
	return 25
}

func (l *ControllerManagerMisconfiguredLab) Tags() []string {
	return []string{"controller-manager", "control-plane", "static-pod", "cluster-architecture"}
}

func (l *ControllerManagerMisconfiguredLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ControllerManagerMisconfiguredLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("getting control plane node: %w", err)
	}
	_, err = dockerExec(ctx, nodeName, "bash", "-c",
		"echo '  --invalid-controller-flag=true' >> /etc/kubernetes/manifests/kube-controller-manager.yaml")
	return err
}

func (l *ControllerManagerMisconfiguredLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-controller-manager",
		"-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return nil
	}

	phase := strings.TrimSpace(output)
	if phase == "Running" {
		return fmt.Errorf("controller manager is running (expected crash)")
	}

	return nil
}

func (l *ControllerManagerMisconfiguredLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-controller-manager",
		"-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return fmt.Errorf("checking controller manager: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("controller manager not running (phase: %s)", strings.TrimSpace(output))
	}

	return nil
}

func (l *ControllerManagerMisconfiguredLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check controller manager pod",
			Command:     "kubectl get pods -n kube-system | grep controller-manager",
			Notes:       "Pod should be in CrashLoopBackOff",
		},
		{
			Description: "Access control plane node",
			Command:     "docker exec -it <control-plane> bash",
			Notes:       "Need to edit the static pod manifest",
		},
		{
			Description: "Remove invalid flag",
			Command:     "sed -i '/--invalid-controller-flag/d' /etc/kubernetes/manifests/kube-controller-manager.yaml",
			Notes:       "Remove the offending flag",
		},
		{
			Description: "Verify controller manager restarts",
			Command:     "kubectl get pods -n kube-system | grep controller-manager",
			Notes:       "Should return to Running state",
		},
	}
}
