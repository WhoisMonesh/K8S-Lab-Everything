package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&PodSchedulingTolerationsLab{})
}

type PodSchedulingTolerationsLab struct {
	BaseLab
}

func (l *PodSchedulingTolerationsLab) ID() string { return "cka_pod_scheduling_tolerations" }
func (l *PodSchedulingTolerationsLab) Title() string {
	return "Schedule Pods with Tolerations"
}
func (l *PodSchedulingTolerationsLab) Category() Category     { return CategoryWorkloadsScheduling }
func (l *PodSchedulingTolerationsLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PodSchedulingTolerationsLab) EstimatedTime() int     { return 20 }
func (l *PodSchedulingTolerationsLab) Tags() []string {
	return []string{"tolerations", "taints", "scheduling"}
}
func (l *PodSchedulingTolerationsLab) Cert() Cert        { return CertCKA }
func (l *PodSchedulingTolerationsLab) DomainWeight() int { return 15 }

func (l *PodSchedulingTolerationsLab) Description() string {
	return `A node has a taint dedicated=dedicated:NoSchedule but pods need to run
on it. Add the appropriate toleration to the pod spec so it can be scheduled
on the tainted node.`
}

func (l *PodSchedulingTolerationsLab) Hints() []string {
	return []string{
		"Check node taints with kubectl describe node",
		"Add tolerations matching the taint key and value",
		"Ensure the toleration effect matches the taint effect",
	}
}

func (l *PodSchedulingTolerationsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodSchedulingTolerationsLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o",
		"jsonpath={.items[0].metadata.name}")
	if err != nil {
		return err
	}
	_, err = kubectl(ctx, kubeconfigPath, "taint", "nodes", strings.TrimSpace(nodeName),
		"dedicated=dedicated:NoSchedule")
	return err
}

func (l *PodSchedulingTolerationsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "tolerations-ns",
		"-o", "jsonpath={.items[*].spec.tolerations}")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "dedicated") {
		return fmt.Errorf("pod does not have correct toleration")
	}
	return nil
}

func (l *PodSchedulingTolerationsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check node taints", Command: "kubectl describe nodes | grep Taints"},
		{Description: "Add toleration to pod", Command: "Add tolerations:\n- key: dedicated\n  value: dedicated\n  effect: NoSchedule"},
		{Description: "Verify pod scheduled", Command: "kubectl get pods -n tolerations-ns -o wide"},
	}
}
