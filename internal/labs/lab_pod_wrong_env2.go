package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodWrongEnv2{})
}

type PodWrongEnv2 struct {
	BaseLab
}

func (l *PodWrongEnv2) ID() string            { return "pod_wrong_env2" }
func (l *PodWrongEnv2) Title() string         { return "Pod Failing Due to Wrong Environment Variable" }
func (l *PodWrongEnv2) Category() Category    { return CategoryWorkloads }
func (l *PodWrongEnv2) Difficulty() Difficulty { return DifficultyEasy }
func (l *PodWrongEnv2) EstimatedTime() int    { return 10 }
func (l *PodWrongEnv2) Tags() []string        { return []string{"environment", "config", "workloads"} }

func (l *PodWrongEnv2) Description() string {
	return `A pod is failing because an environment variable has the wrong value.
Fix the environment variable to the correct value.`
}

func (l *PodWrongEnv2) Hints() []string {
	return []string{
		"Check the pod environment variables",
		"Look at the env configuration",
		"Fix the value to match expected",
	}
}

func (l *PodWrongEnv2) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodWrongEnv2) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: env-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: env-app
  template:
    metadata:
      labels:
        app: env-app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ["sh", "-c", "if [ $APP_MODE != 'production' ]; then exit 1; fi; sleep 3600"]
        env:
        - name: APP_MODE
          value: "staging"`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PodWrongEnv2) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/env-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return err
	}
	if output != "1" {
		return fmt.Errorf("deployment not ready")
	}
	return nil
}

func (l *PodWrongEnv2) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check env vars", Command: "kubectl get deploy env-app -o jsonpath='{.spec.template.spec.containers[0].env}'"},
		{Description: "Fix env value", Command: "kubectl patch deploy env-app -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"app\",\"env\":[{\"name\":\"APP_MODE\",\"value\":\"production\"}]}]}}}}'"},
	}
}
