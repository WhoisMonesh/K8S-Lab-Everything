package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&PriorityClassMissingLab{}) }

type PriorityClassMissingLab struct{ BaseLab }

func (l *PriorityClassMissingLab) ID() string             { return "priorityclass_missing" }
func (l *PriorityClassMissingLab) Title() string          { return "Pod Uses Nonexistent PriorityClass" }
func (l *PriorityClassMissingLab) Category() Category     { return CategoryScheduling }
func (l *PriorityClassMissingLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PriorityClassMissingLab) EstimatedTime() int     { return 15 }
func (l *PriorityClassMissingLab) Tags() []string {
	return []string{"priorityclass", "scheduling", "pending"}
}
func (l *PriorityClassMissingLab) Description() string {
	return `A pod 'critical-task' is stuck Pending. The pod spec references a
PriorityClass named 'high-priority' which doesn't exist in the
cluster.

Your task: Create the missing PriorityClass (value=1000000,
preemptionPolicy=Never) and verify the pod gets scheduled.`
}
func (l *PriorityClassMissingLab) Hints() []string {
	return []string{
		"Check: kubectl describe pod critical-task — look for error about PriorityClass",
		"Verify: kubectl get priorityclass — 'high-priority' is missing",
		"Create: kubectl create priorityclass high-priority --value=1000000 --preemption-policy=Never",
	}
}

func (l *PriorityClassMissingLab) Break(ctx context.Context, kp string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: critical-task
  namespace: default
spec:
  priorityClassName: high-priority
  containers:
  - name: work
    image: busybox:1.36
    command: ["sh","-c","echo running critical work; sleep 999"]
`
	return kubectlApply(ctx, kp, pod)
}

func (l *PriorityClassMissingLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(3 * time.Second)
	return nil
}

func (l *PriorityClassMissingLab) Verify(ctx context.Context, kp string) error {
	pc, _ := kubectl(ctx, kp, "get", "priorityclass", "high-priority", "-o", "jsonpath={.value}")
	if pc == "" {
		return fmt.Errorf("PriorityClass 'high-priority' still doesn't exist")
	}
	phase, _ := kubectl(ctx, kp, "get", "pod", "critical-task", "-o", "jsonpath={.status.phase}")
	if phase != "Running" {
		return fmt.Errorf("pod not running (phase: %s)", phase)
	}
	return nil
}

func (l *PriorityClassMissingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod events", Command: "kubectl describe pod critical-task | tail -10", Notes: "Error: nonexistent PriorityClass 'high-priority'"},
		{Description: "Create the PriorityClass", Command: "kubectl create priorityclass high-priority --value=1000000 --preemption-policy=Never", Notes: "Now the pod can reference it"},
		{Description: "Verify pod schedules", Command: "kubectl get pod critical-task", Notes: "Pod moves from Pending to Running"},
	}
}
