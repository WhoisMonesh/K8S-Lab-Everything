package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&OOMKilledLimitsLab{})
}

type OOMKilledLimitsLab struct {
	BaseLab
}

func (l *OOMKilledLimitsLab) ID() string {
	return "oomkilled_limits"
}

func (l *OOMKilledLimitsLab) Title() string {
	return "Pods OOMKilled By Memory Limits"
}

func (l *OOMKilledLimitsLab) Category() Category {
	return CategoryWorkloads
}

func (l *OOMKilledLimitsLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *OOMKilledLimitsLab) Description() string {
	return `A data processing deployment named 'processor' keeps restarting.
The pods flip between Running and CrashLoopBackOff, and kubectl describe pod
shows the last state as OOMKilled (exit code 137).

Your task: Fix the resource configuration so both replicas stay running.`
}

func (l *OOMKilledLimitsLab) Hints() []string {
	return []string{
		"kubectl describe pod shows 'Last State: Terminated, Reason: OOMKilled'",
		"Check the memory limit on the container spec",
		"The application intentionally allocates more than 16Mi of memory",
		"Increase the limits and requests to something the app can live with",
	}
}

func (l *OOMKilledLimitsLab) EstimatedTime() int {
	return 15
}

func (l *OOMKilledLimitsLab) Tags() []string {
	return []string{"resources", "limits", "oomkilled", "memory", "workloads"}
}

func (l *OOMKilledLimitsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *OOMKilledLimitsLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: processor
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: processor
  template:
    metadata:
      labels:
        app: processor
    spec:
      containers:
      - name: worker
        image: polinux/stress:1.0.4
        command: ["stress"]
        args: ["--vm", "1", "--vm-bytes", "64M", "--vm-hang", "0", "--timeout", "3600s"]
        resources:
          limits:
            memory: "16Mi"
            cpu: "200m"
          requests:
            memory: "8Mi"
            cpu: "100m"
`

	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating broken deployment: %w", err)
	}
	return nil
}

func (l *OOMKilledLimitsLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(20 * time.Second)
	return nil
}

func (l *OOMKilledLimitsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "processor",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}
	if output != "2" {
		return fmt.Errorf("deployment not fully ready yet (ready replicas: %s, expected: 2)", output)
	}

	pods, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=processor",
		"-o", "jsonpath={.items[*].status.containerStatuses[*].restartCount}")
	if err != nil {
		return fmt.Errorf("failed to check pods: %w", err)
	}
	for _, count := range splitFields(pods) {
		if count == "0" {
			return nil
		}
	}
	return fmt.Errorf("all pods have restarts - raise limits high enough that a fresh pod stays up")
}

func (l *OOMKilledLimitsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Look at why pods are restarting",
			Command:     "kubectl describe pod -l app=processor | grep -A 5 'Last State'",
			Notes:       "You will see Reason: OOMKilled - the kernel killed the container for exceeding its memory limit",
		},
		{
			Description: "Review the configured limits",
			Command:     "kubectl get deploy processor -o jsonpath='{.spec.template.spec.containers[0].resources}'",
			Notes:       "The memory limit is only 16Mi while the workload allocates 64M",
		},
		{
			Description: "Raise the memory limits",
			Command:     "kubectl set resources deploy/processor --limits=memory=128Mi,cpu=500m --requests=memory=64Mi,cpu=100m",
			Notes:       "Any value comfortably above the 64MB working set works",
		},
		{
			Description: "Wait for the rollout and confirm stability",
			Command:     "kubectl rollout status deploy/processor && kubectl get pods -l app=processor",
			Notes:       "Both replicas should be Running with no further restarts",
		},
	}
}
