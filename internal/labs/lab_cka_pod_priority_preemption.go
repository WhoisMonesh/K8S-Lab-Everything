package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodPriorityPreemptionLab{})
}

type PodPriorityPreemptionLab struct {
	BaseLab
}

func (l *PodPriorityPreemptionLab) ID() string { return "cka_pod_priority_preemption" }
func (l *PodPriorityPreemptionLab) Title() string {
	return "Configure Pod Priority Classes"
}
func (l *PodPriorityPreemptionLab) Category() Category     { return CategoryWorkloadsScheduling }
func (l *PodPriorityPreemptionLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PodPriorityPreemptionLab) EstimatedTime() int     { return 20 }
func (l *PodPriorityPreemptionLab) Tags() []string {
	return []string{"priority", "preemption", "scheduling"}
}
func (l *PodPriorityPreemptionLab) Cert() Cert        { return CertCKA }
func (l *PodPriorityPreemptionLab) DomainWeight() int { return 15 }

func (l *PodPriorityPreemptionLab) Description() string {
	return `Create a high-priority PriorityClass and deploy a pod that uses it.
The high-priority pod should be able to preempt lower-priority pods when
cluster resources are constrained.`
}

func (l *PodPriorityPreemptionLab) Hints() []string {
	return []string{
		"Create a PriorityClass with high value",
		"Set preemptionPolicy to PreemptLowerPriority",
		"Reference the PriorityClass in pod spec",
	}
}

func (l *PodPriorityPreemptionLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodPriorityPreemptionLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *PodPriorityPreemptionLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "priorityclass", "high-priority",
		"-o", "jsonpath={.value}")
	if err != nil {
		return err
	}
	if output != "1000000" {
		return fmt.Errorf("PriorityClass value not set to 1000000")
	}
	return nil
}

func (l *PodPriorityPreemptionLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create PriorityClass", Command: "cat <<EOF | kubectl apply -f -\napiVersion: scheduling.k8s.io/v1\nkind: PriorityClass\nmetadata:\n  name: high-priority\nvalue: 1000000\nglobalDefault: false\npreemptionPolicy: PreemptLowerPriority\ndescription: High priority class\nEOF"},
		{Description: "Create high-priority pod", Command: "kubectl run high-pod --image=nginx --restart=Never --overrides='{\"spec\":{\"priorityClassName\":\"high-priority\"}}'"},
		{Description: "Verify", Command: "kubectl get pods -o wide"},
	}
}
