package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PriorityClassPreemptionLab{})
}

type PriorityClassPreemptionLab struct {
	BaseLab
}

func (l *PriorityClassPreemptionLab) ID() string {
	return "priority_class_preemption"
}

func (l *PriorityClassPreemptionLab) Title() string {
	return "High Priority Pod Preempting"
}

func (l *PriorityClassPreemptionLab) Category() Category {
	return CategoryScheduling
}

func (l *PriorityClassPreemptionLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *PriorityClassPreemptionLab) Description() string {
	return `A high-priority Pod 'critical-task' is preempting lower-priority
'batch-worker' pods. The batch workers keep getting evicted when the
critical task is scheduled.

Your task: Configure preemption settings to protect the batch workers.`
}

func (l *PriorityClassPreemptionLab) Hints() []string {
	return []string{
		"Check PriorityClass settings",
		"preemptionPolicy controls whether pods can preempt others",
		"Set preemptionPolicy to Never for the batch PriorityClass",
	}
}

func (l *PriorityClassPreemptionLab) EstimatedTime() int {
	return 20
}

func (l *PriorityClassPreemptionLab) Tags() []string {
	return []string{"priority", "preemption", "scheduling"}
}

func (l *PriorityClassPreemptionLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PriorityClassPreemptionLab) Break(ctx context.Context, kubeconfigPath string) error {
	lowPriority := `apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: batch-low
value: 100
globalDefault: false
preemptionPolicy: PreemptLowerPriority
description: "Low priority for batch workloads"
`
	if err := kubectlApply(ctx, kubeconfigPath, lowPriority); err != nil {
		return fmt.Errorf("creating low priority class: %w", err)
	}

	highPriority := `apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: critical-high
value: 1000
globalDefault: false
preemptionPolicy: PreemptLowerPriority
description: "High priority for critical workloads"
`
	if err := kubectlApply(ctx, kubeconfigPath, highPriority); err != nil {
		return fmt.Errorf("creating high priority class: %w", err)
	}

	batchWorker := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: batch-worker
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: batch-worker
  template:
    metadata:
      labels:
        app: batch-worker
    spec:
      priorityClassName: batch-low
      containers:
      - name: worker
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo working; sleep 15; done']
`
	if err := kubectlApply(ctx, kubeconfigPath, batchWorker); err != nil {
		return fmt.Errorf("creating batch worker: %w", err)
	}

	time.Sleep(10 * time.Second)

	criticalTask := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: critical-task
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: critical-task
  template:
    metadata:
      labels:
        app: critical-task
    spec:
      priorityClassName: critical-high
      containers:
      - name: critical
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo critical; sleep 15; done']
`
	if err := kubectlApply(ctx, kubeconfigPath, criticalTask); err != nil {
		return fmt.Errorf("creating critical task: %w", err)
	}

	return nil
}

func (l *PriorityClassPreemptionLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *PriorityClassPreemptionLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "priorityclass", "batch-low",
		"-o", "jsonpath={.preemptionPolicy}")
	if err != nil {
		return fmt.Errorf("failed to check priority class: %w", err)
	}

	if strings.TrimSpace(output) == "PreemptLowerPriority" {
		return fmt.Errorf("batch-low priority class still allows preemption")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=batch-worker",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check batch pods: %w", err)
	}

	runningCount := 0
	for _, phase := range splitFields(output) {
		if phase == "Running" {
			runningCount++
		}
	}

	if runningCount < 2 {
		return fmt.Errorf("not enough batch workers running (running: %d)", runningCount)
	}

	return nil
}

func (l *PriorityClassPreemptionLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check PriorityClasses",
			Command:     "kubectl get priorityclasses",
			Notes:       "batch-low has preemptionPolicy: PreemptLowerPriority",
		},
		{
			Description: "Check batch worker pods",
			Command:     "kubectl get pods -l app=batch-worker",
			Notes:       "Some pods may have been preempted",
		},
		{
			Description: "Fix batch PriorityClass",
			Command:     "kubectl patch priorityclass batch-low --type='json' -p='[{\"op\":\"replace\",\"path\":\"/preemptionPolicy\",\"value\":\"Never\"}]'",
			Notes:       "Set preemptionPolicy to Never to protect batch workloads",
		},
		{
			Description: "Verify batch workers are protected",
			Command:     "kubectl get pods -l app=batch-worker",
			Notes:       "All 3 batch workers should now be running",
		},
	}
}
