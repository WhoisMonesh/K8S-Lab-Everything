package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&SecretEnvBrokenLab{})
}

type SecretEnvBrokenLab struct {
	BaseLab
}

func (l *SecretEnvBrokenLab) ID() string {
	return "secret_env_broken"
}

func (l *SecretEnvBrokenLab) Title() string {
	return "Pod Failing Due To Broken Secret Reference"
}

func (l *SecretEnvBrokenLab) Category() Category {
	return CategorySecurity
}

func (l *SecretEnvBrokenLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *SecretEnvBrokenLab) Description() string {
	return `An application pod 'report-gen' cannot start. It stays in
CreateContainerConfigError state.

The pod consumes credentials from a Secret via environment variables, but
something about that Secret is wrong.

Your task: Investigate and fix so the pod starts successfully.`
}

func (l *SecretEnvBrokenLab) Hints() []string {
	return []string{
		"kubectl describe pod report-gen shows the exact error event",
		"List the keys inside the referenced secret with kubectl get secret ... -o yaml",
		"The pod asks for an env var sourced from a specific key in the secret",
		"Either add the missing key to the secret or repoint the pod at the key that exists",
	}
}

func (l *SecretEnvBrokenLab) EstimatedTime() int {
	return 15
}

func (l *SecretEnvBrokenLab) Tags() []string {
	return []string{"secret", "env", "configuration", "security", "troubleshooting"}
}

func (l *SecretEnvBrokenLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SecretEnvBrokenLab) Break(ctx context.Context, kubeconfigPath string) error {
	secret := `apiVersion: v1
kind: Secret
metadata:
  name: report-credentials
  namespace: default
type: Opaque
stringData:
  api_key: super-secret-key-123
`

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: report-gen
  namespace: default
spec:
  restartPolicy: Never
  containers:
  - name: reporter
    image: busybox:1.36
    command: ["sh", "-c", "echo using key=$API_KEY && sleep 3600"]
    env:
    - name: API_KEY
      valueFrom:
        secretKeyRef:
          name: report-credentials
          key: apikey
`

	for _, manifest := range []string{secret, pod} {
		if err := kubectlApply(ctx, kubeconfigPath, manifest); err != nil {
			return fmt.Errorf("applying lab resources: %w", err)
		}
	}
	return nil
}

func (l *SecretEnvBrokenLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(8 * time.Second)
	return nil
}

func (l *SecretEnvBrokenLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "report-gen",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}
	if output != "Running" {
		return fmt.Errorf("pod report-gen is not running yet (phase: %s)", output)
	}

	logs, _ := kubectl(ctx, kubeconfigPath, "logs", "report-gen")
	if logs == "" {
		return fmt.Errorf("pod has not produced output yet, wait a few seconds")
	}
	return nil
}

func (l *SecretEnvBrokenLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Inspect the pod events",
			Command:     "kubectl describe pod report-gen",
			Notes:       "Events show: key 'apikey' not found in secret 'report-credentials'",
		},
		{
			Description: "Look at what the secret actually contains",
			Command:     "kubectl get secret report-credentials -o yaml",
			Notes:       "The secret only defines a key named api_key, not apikey",
		},
		{
			Description: "Fix option A: patch the pod to use the existing key",
			Command:     "kubectl get pod report-gen -o yaml > report-gen.yaml   # then change secretKeyRef.key to api_key and re-apply",
			Notes:       "Pods are immutable for env values, so recreate the pod after editing",
		},
		{
			Description: "Fix option B: add the expected key to the secret",
			Command:     "kubectl patch secret report-credentials -p '{\"stringData\":{\"apikey\":\"super-secret-key-123\"}}'",
			Notes:       "The kubelet will retry mounting env vars and the pod will start",
		},
		{
			Description: "Verify the pod is running",
			Command:     "kubectl get pod report-gen && kubectl logs report-gen",
			Notes:       "Logs should print 'using key=super-secret-key-123'",
		},
	}
}
