package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ConfigMapWrongKeyLab{})
}

type ConfigMapWrongKeyLab struct {
	BaseLab
}

func (l *ConfigMapWrongKeyLab) ID() string {
	return "configmap_wrong_key"
}

func (l *ConfigMapWrongKeyLab) Title() string {
	return "ConfigMap Key Mismatch"
}

func (l *ConfigMapWrongKeyLab) Category() Category {
	return CategoryWorkloads
}

func (l *ConfigMapWrongKeyLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *ConfigMapWrongKeyLab) Description() string {
	return `A pod 'config-app' is failing to start because it references a ConfigMap key that doesn't exist.
The ConfigMap exists but the key name used in the pod spec is wrong.

Your task: Fix the pod configuration to use the correct ConfigMap key.`
}

func (l *ConfigMapWrongKeyLab) Hints() []string {
	return []string{
		"Check the pod status and events",
		"Look at the pod's env or volume mount configuration",
		"Check the ConfigMap keys",
		"The key name in the pod spec doesn't match the ConfigMap data key",
	}
}

func (l *ConfigMapWrongKeyLab) EstimatedTime() int {
	return 10
}

func (l *ConfigMapWrongKeyLab) Tags() []string {
	return []string{"configmap", "secrets", "environment", "troubleshooting"}
}

func (l *ConfigMapWrongKeyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ConfigMapWrongKeyLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create ConfigMap with correct key
	configMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  APP_MODE: "production"
  LOG_LEVEL: "info"
`
	if err := kubectlApply(ctx, kubeconfigPath, configMap); err != nil {
		return fmt.Errorf("creating ConfigMap: %w", err)
	}

	// Create pod referencing wrong key name
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: config-app
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.28
    command: ['sh', '-c', 'echo "Mode: $APP_MODE" && sleep 3600']
    env:
    - name: APP_MODE
      valueFrom:
        configMapKeyRef:
          name: app-config
          key: app_mode
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *ConfigMapWrongKeyLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ConfigMapWrongKeyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if pod is running
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "config-app",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	// Check environment variable
	output, err = kubectl(ctx, kubeconfigPath, "exec", "config-app",
		"--", "printenv", "APP_MODE")
	if err != nil {
		return fmt.Errorf("failed to exec into pod: %w", err)
	}

	if strings.TrimSpace(output) != "production" {
		return fmt.Errorf("APP_MODE env var not set correctly (got: %s)", output)
	}

	return nil
}

func (l *ConfigMapWrongKeyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod config-app",
			Notes:       "The pod should be in CreateError or CrashLoopBackOff",
		},
		{
			Description: "Check pod events",
			Command:     "kubectl describe pod config-app | grep -A 10 Events",
			Notes:       "Look for error about ConfigMap key not found",
		},
		{
			Description: "Check the ConfigMap keys",
			Command:     "kubectl get configmap app-config -o yaml",
			Notes:       "The key is 'APP_MODE' (uppercase with underscore)",
		},
		{
			Description: "Check the pod spec",
			Command:     "kubectl get pod config-app -o yaml | grep -A 5 env",
			Notes:       "The pod references key 'app_mode' (lowercase with underscore) which doesn't exist",
		},
		{
			Description: "Fix the ConfigMap key reference in the pod",
			Command:     "kubectl edit pod config-app",
			Notes:       "Change key from 'app_mode' to 'APP_MODE' to match the ConfigMap",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod config-app",
			Notes:       "The pod should now be in Running state",
		},
	}
}
