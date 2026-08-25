package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&EnvVarMissingLab{})
}

type EnvVarMissingLab struct {
	BaseLab
}

func (l *EnvVarMissingLab) ID() string {
	return "env_var_missing"
}

func (l *EnvVarMissingLab) Title() string {
	return "Pod Failing Due to Missing Environment Variable"
}

func (l *EnvVarMissingLab) Category() Category {
	return CategoryWorkloads
}

func (l *EnvVarMissingLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *EnvVarMissingLab) Description() string {
	return `A pod 'worker' is crashing because it expects an environment variable 'DATABASE_URL' that is not set.

Your task: Fix the pod configuration to provide the required environment variable.`
}

func (l *EnvVarMissingLab) Hints() []string {
	return []string{
		"Check the pod logs",
		"Look at the pod's environment variables",
		"The pod expects DATABASE_URL but it's not defined",
		"Add the missing env var to the pod spec",
	}
}

func (l *EnvVarMissingLab) EstimatedTime() int {
	return 10
}

func (l *EnvVarMissingLab) Tags() []string {
	return []string{"environment", "configmap", "env", "troubleshooting"}
}

func (l *EnvVarMissingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *EnvVarMissingLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create pod that requires DATABASE_URL env var but doesn't define it
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: worker
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.28
    command: ['sh', '-c', 'if [ -z "$DATABASE_URL" ]; then echo "Error: DATABASE_URL is required" && exit 1; fi && echo "Connected to: $DATABASE_URL" && sleep 3600']
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *EnvVarMissingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *EnvVarMissingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if pod is running
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "worker",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	// Check environment variable
	output, err = kubectl(ctx, kubeconfigPath, "exec", "worker",
		"--", "printenv", "DATABASE_URL")
	if err != nil {
		return fmt.Errorf("failed to exec into pod: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("DATABASE_URL env var not set")
	}

	return nil
}

func (l *EnvVarMissingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod worker",
			Notes:       "The pod should be in Error or CrashLoopBackOff",
		},
		{
			Description: "Check pod logs",
			Command:     "kubectl logs worker",
			Notes:       "Logs show 'DATABASE_URL is required' error",
		},
		{
			Description: "Check pod environment variables",
			Command:     "kubectl get pod worker -o yaml | grep -A 10 env",
			Notes:       "The pod spec has no env section defined",
		},
		{
			Description: "Delete and recreate pod with env var",
			Command:     "kubectl delete pod worker",
			Notes:       "Delete the broken pod first",
		},
		{
			Description: "Create pod with DATABASE_URL",
			Command:     `kubectl run worker --image=busybox:1.28 --env="DATABASE_URL=postgres://db:5432/mydb" -- sleep 3600`,
			Notes:       "Or edit the pod YAML to add the env var",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod worker",
			Notes:       "The pod should now be in Running state",
		},
	}
}
