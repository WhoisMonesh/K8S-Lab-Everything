package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&SecretVolumeMissingLab{})
}

type SecretVolumeMissingLab struct {
	BaseLab
}

func (l *SecretVolumeMissingLab) ID() string {
	return "secret_volume_missing"
}

func (l *SecretVolumeMissingLab) Title() string {
	return "Secret Volume Not Found"
}

func (l *SecretVolumeMissingLab) Category() Category {
	return CategoryWorkloads
}

func (l *SecretVolumeMissingLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *SecretVolumeMissingLab) Description() string {
	return `A pod 'secure-app' is failing because it references a Secret volume that doesn't exist.
The pod spec references a Secret named 'app-tls' but the actual Secret has a different name.

Your task: Fix the pod configuration to reference the correct Secret.`
}

func (l *SecretVolumeMissingLab) Hints() []string {
	return []string{
		"Check pod events for Secret not found errors",
		"List available Secrets in the namespace",
		"Compare the Secret name in the pod spec with actual Secrets",
	}
}

func (l *SecretVolumeMissingLab) EstimatedTime() int {
	return 10
}

func (l *SecretVolumeMissingLab) Tags() []string {
	return []string{"secret", "volume", "troubleshooting"}
}

func (l *SecretVolumeMissingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SecretVolumeMissingLab) Break(ctx context.Context, kubeconfigPath string) error {
	secret := `apiVersion: v1
kind: Secret
metadata:
  name: app-certs
  namespace: default
type: Opaque
data:
  tls.crt: LS0tLS1CRUdJTi...fake
  tls.key: LS0tLS1CRUdJTi...fake
`
	if err := kubectlApply(ctx, kubeconfigPath, secret); err != nil {
		return fmt.Errorf("creating Secret: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: secure-app
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'ls /etc/certs && sleep 3600']
    volumeMounts:
    - name: tls-vol
      mountPath: /etc/certs
      readOnly: true
  volumes:
  - name: tls-vol
    secret:
      secretName: app-tls
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *SecretVolumeMissingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *SecretVolumeMissingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "secure-app",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	output, err = kubectl(ctx, kubeconfigPath, "exec", "secure-app",
		"--", "ls", "/etc/certs")
	if err != nil {
		return fmt.Errorf("cannot access secret volume: %w", err)
	}

	if !strings.Contains(output, "tls.crt") {
		return fmt.Errorf("tls.crt not found in secret volume")
	}

	return nil
}

func (l *SecretVolumeMissingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod secure-app",
			Notes:       "Pod should be in CreateError state",
		},
		{
			Description: "Check pod events",
			Command:     "kubectl describe pod secure-app | grep -A 10 Events",
			Notes:       "Look for Secret 'app-tls' not found error",
		},
		{
			Description: "List available Secrets",
			Command:     "kubectl get secrets",
			Notes:       "The actual Secret is named 'app-certs'",
		},
		{
			Description: "Fix the Secret reference",
			Command:     "kubectl edit pod secure-app",
			Notes:       "Change secretName from 'app-tls' to 'app-certs'",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod secure-app",
			Notes:       "Pod should be Running with secret mounted",
		},
	}
}
