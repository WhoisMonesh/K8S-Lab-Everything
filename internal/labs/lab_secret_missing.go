package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&SecretMissingLab{})
}

type SecretMissingLab struct {
	BaseLab
}

func (l *SecretMissingLab) ID() string {
	return "secret_missing"
}

func (l *SecretMissingLab) Title() string {
	return "Pod Failing Due to Missing Secret"
}

func (l *SecretMissingLab) Category() Category {
	return CategorySecurity
}

func (l *SecretMissingLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *SecretMissingLab) Description() string {
	return `A pod 'db-connector' in the 'database' namespace is failing because it references a Secret that doesn't exist.
The pod needs database credentials from a Secret named 'db-credentials'.

Your task: Create the missing Secret with the correct data.`
}

func (l *SecretMissingLab) Hints() []string {
	return []string{
		"Check the pod status and events",
		"Look at the pod's environment variable configuration",
		"Check if the Secret 'db-credentials' exists",
		"Create the Secret with the required keys",
	}
}

func (l *SecretMissingLab) EstimatedTime() int {
	return 10
}

func (l *SecretMissingLab) Tags() []string {
	return []string{"secrets", "security", "environment", "troubleshooting"}
}

func (l *SecretMissingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SecretMissingLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create namespace
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: database
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	// Create pod that references non-existent Secret
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: db-connector
  namespace: database
spec:
  containers:
  - name: app
    image: busybox:1.28
    command: ['sh', '-c', 'echo "DB: $DB_HOST" && sleep 3600']
    env:
    - name: DB_HOST
      valueFrom:
        secretKeyRef:
          name: db-credentials
          key: host
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *SecretMissingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *SecretMissingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if pod is running
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "db-connector", "-n", "database",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	// Check environment variable
	output, err = kubectl(ctx, kubeconfigPath, "exec", "-n", "database", "db-connector",
		"--", "printenv", "DB_HOST")
	if err != nil {
		return fmt.Errorf("failed to exec into pod: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("DB_HOST env var not set")
	}

	return nil
}

func (l *SecretMissingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod db-connector -n database",
			Notes:       "The pod should be in CreateError or CrashLoopBackOff",
		},
		{
			Description: "Check pod events",
			Command:     "kubectl describe pod db-connector -n database | grep -A 10 Events",
			Notes:       "Look for error about Secret 'db-credentials' not found",
		},
		{
			Description: "Check if Secret exists",
			Command:     "kubectl get secret db-credentials -n database",
			Notes:       "The Secret does not exist in the namespace",
		},
		{
			Description: "Create the missing Secret",
			Command:     `kubectl create secret generic db-credentials -n database --from-literal=host=mysql.default.svc.cluster.local --from-literal=username=admin --from-literal=password=secret123`,
			Notes:       "Create the Secret with the required database credentials",
		},
		{
			Description: "Restart the pod to pick up the new Secret",
			Command:     "kubectl delete pod db-connector -n database && kubectl apply -f pod.yaml",
			Notes:       "Delete and recreate the pod to load the new Secret",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod db-connector -n database",
			Notes:       "The pod should now be in Running state",
		},
	}
}
