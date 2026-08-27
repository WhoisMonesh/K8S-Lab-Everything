package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PrestopHookWrong{})
}

type PrestopHookWrong struct {
	BaseLab
}

func (l *PrestopHookWrong) ID() string             { return "prestop_hook_wrong" }
func (l *PrestopHookWrong) Title() string          { return "PreStop Hook Causing Pod Termination Issues" }
func (l *PrestopHookWrong) Category() Category     { return CategoryWorkloads }
func (l *PrestopHookWrong) Difficulty() Difficulty { return DifficultyMedium }
func (l *PrestopHookWrong) EstimatedTime() int     { return 15 }
func (l *PrestopHookWrong) Tags() []string         { return []string{"prestop", "lifecycle", "workloads"} }

func (l *PrestopHookWrong) Description() string {
	return `A pod is taking too long to terminate because the preStop hook has an incorrect command.
Fix the preStop hook configuration.`
}

func (l *PrestopHookWrong) Hints() []string {
	return []string{
		"Check the pod lifecycle configuration",
		"Look at the preStop hook command",
		"Verify the command exists in the container image",
	}
}

func (l *PrestopHookWrong) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PrestopHookWrong) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: prestop-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: prestop-app
  template:
    metadata:
      labels:
        app: prestop-app
    spec:
      terminationGracePeriodSeconds: 30
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
        lifecycle:
          preStop:
            exec:
              command: ["/bin/sh", "-c", "sleep 60"]`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PrestopHookWrong) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/prestop-app",
		"-o", "jsonpath={.spec.template.spec.containers[0].lifecycle.preStop.exec.command[2]}")
	if err != nil {
		return err
	}
	if output == "sleep 60" {
		return fmt.Errorf("preStop hook still has long sleep")
	}
	return nil
}

func (l *PrestopHookWrong) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check preStop hook", Command: "kubectl get deploy prestop-app -o jsonpath='{.spec.template.spec.containers[0].lifecycle}'"},
		{Description: "Fix preStop hook", Command: "kubectl edit deploy prestop-app"},
		{Description: "Change sleep to shorter duration", Command: "Change sleep 60 to sleep 5 or remove the preStop hook"},
	}
}
