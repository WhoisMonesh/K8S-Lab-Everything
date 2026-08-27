package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodPreemptionOccurred{})
}

type PodPreemptionOccurred struct {
	BaseLab
}

func (l *PodPreemptionOccurred) ID() string             { return "pod_preemption_occurred" }
func (l *PodPreemptionOccurred) Title() string          { return "Low Priority Pod Preempted" }
func (l *PodPreemptionOccurred) Category() Category     { return CategoryScheduling }
func (l *PodPreemptionOccurred) Difficulty() Difficulty { return DifficultyMedium }
func (l *PodPreemptionOccurred) EstimatedTime() int     { return 15 }
func (l *PodPreemptionOccurred) Tags() []string {
	return []string{"scheduling", "priority", "preemption"}
}

func (l *PodPreemptionOccurred) Description() string {
	return `A low priority pod was preempted by a higher priority pod.
Understand the preemption and ensure the critical pod runs correctly.`
}

func (l *PodPreemptionOccurred) Hints() []string {
	return []string{
		"Check pod priority classes",
		"Look at pod events for preemption messages",
		"Verify which pods have higher priority",
	}
}

func (l *PodPreemptionOccurred) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodPreemptionOccurred) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 1000000
globalDefault: false
preemptionPolicy: PreemptLowerPriority
---
apiVersion: v1
kind: Pod
metadata:
  name: low-priority-pod
spec:
  priorityClassName: default
  containers:
  - name: nginx
    image: nginx:alpine
---
apiVersion: v1
kind: Pod
metadata:
  name: high-priority-pod
spec:
  priorityClassName: high-priority
  containers:
  - name: nginx
    image: nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PodPreemptionOccurred) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "high-priority-pod",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("high priority pod not running: %s", output)
	}
	return nil
}

func (l *PodPreemptionOccurred) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod priorities", Command: "kubectl get pods -o custom-columns=NAME:.metadata.name,PRIORITY:.spec.priorityClassName"},
		{Description: "Check events", Command: "kubectl get events --field-selector reason=Preempted"},
		{Description: "Verify high priority pod runs", Command: "kubectl get pod high-priority-pod"},
	}
}
