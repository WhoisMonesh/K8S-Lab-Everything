package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&SchedulerNotSchedulingLab{})
}

type SchedulerNotSchedulingLab struct {
	BaseLab
}

func (l *SchedulerNotSchedulingLab) ID() string { return "cka_scheduler_not_scheduling" }
func (l *SchedulerNotSchedulingLab) Title() string {
	return "Debug Scheduler Issues"
}
func (l *SchedulerNotSchedulingLab) Category() Category     { return CategoryTroubleshooting }
func (l *SchedulerNotSchedulingLab) Difficulty() Difficulty { return DifficultyHard }
func (l *SchedulerNotSchedulingLab) EstimatedTime() int     { return 25 }
func (l *SchedulerNotSchedulingLab) Tags() []string {
	return []string{"scheduler", "scheduling", "troubleshooting"}
}
func (l *SchedulerNotSchedulingLab) Cert() Cert        { return CertCKA }
func (l *SchedulerNotSchedulingLab) DomainWeight() int { return 30 }

func (l *SchedulerNotSchedulingLab) Description() string {
	return `The scheduler is not scheduling new pods. Pods remain in Pending state.
Debug the scheduler by checking its logs and health status.`
}

func (l *SchedulerNotSchedulingLab) Hints() []string {
	return []string{
		"Check scheduler pod status",
		"Review scheduler logs",
		"Verify scheduler configuration",
	}
}

func (l *SchedulerNotSchedulingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SchedulerNotSchedulingLab) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: sched-ns
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sched-app
  namespace: sched-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sched-app
  template:
    metadata:
      labels:
        app: sched-app
    spec:
      nodeName: non-existent-node
      containers:
      - name: app
        image: nginx:1.27-alpine
`
	if err := kubectlApply(ctx, kubeconfigPath, manifest); err != nil {
		return fmt.Errorf("creating broken deployment: %w", err)
	}
	return nil
}

func (l *SchedulerNotSchedulingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *SchedulerNotSchedulingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "sched-ns",
		"-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return err
	}
	if output == "Pending" {
		return fmt.Errorf("pods still pending scheduling")
	}
	return nil
}

func (l *SchedulerNotSchedulingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check scheduler", Command: "kubectl get pods -n kube-system -l component=kube-scheduler"},
		{Description: "Check logs", Command: "kubectl logs -n kube-system -l component=kube-scheduler --tail=50"},
		{Description: "Verify health", Command: "kubectl get --raw /healthz/poststarthook/scheduling/filters"},
		{Description: "Restart scheduler", Command: "Delete scheduler pod to restart"},
	}
}
