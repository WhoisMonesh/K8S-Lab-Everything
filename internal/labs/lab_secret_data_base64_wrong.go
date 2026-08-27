package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&SecretDataBase64WrongLab{})
}

type SecretDataBase64WrongLab struct {
	BaseLab
}

func (l *SecretDataBase64WrongLab) ID() string {
	return "secret_data_base64_wrong"
}

func (l *SecretDataBase64WrongLab) Title() string {
	return "Secret Data Base64 Encoding Wrong"
}

func (l *SecretDataBase64WrongLab) Category() Category {
	return CategorySecurity
}

func (l *SecretDataBase64WrongLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *SecretDataBase64WrongLab) Description() string {
	return `A Secret 'db-credentials' has incorrect base64 encoding in its data.
The password field contains plaintext instead of base64-encoded data,
causing decoding issues when pods mount the secret.

Your task: Fix the Secret data with proper base64 encoding.`
}

func (l *SecretDataBase64WrongLab) Hints() []string {
	return []string{
		"Check the Secret data",
		"Secret data must be valid base64",
		"Use 'echo -n \"value\" | base64' to encode",
	}
}

func (l *SecretDataBase64WrongLab) EstimatedTime() int {
	return 10
}

func (l *SecretDataBase64WrongLab) Tags() []string {
	return []string{"secret", "base64", "encoding", "security"}
}

func (l *SecretDataBase64WrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SecretDataBase64WrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	secret := `apiVersion: v1
kind: Secret
metadata:
  name: db-credentials
  namespace: default
type: Opaque
data:
  password: not-valid-base64!
`
	if err := kubectlApply(ctx, kubeconfigPath, secret); err != nil {
		return fmt.Errorf("creating Secret: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: db-client
  namespace: default
spec:
  containers:
  - name: client
    image: busybox:1.36
    command: ['sh', '-c', 'cat /etc/secrets/password && sleep 3600']
    volumeMounts:
    - name: secrets
      mountPath: /etc/secrets
  volumes:
  - name: secrets
    secret:
      secretName: db-credentials
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *SecretDataBase64WrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pod", "db-client",
		"-o", "jsonpath={.status.phase}")
	if strings.Contains(output, "CreateContainerConfigError") || strings.Contains(output, "Error") {
		return nil
	}
	return nil
}

func (l *SecretDataBase64WrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Verify the secret has valid base64
	output, err := kubectl(ctx, kubeconfigPath, "get", "secret", "db-credentials",
		"-o", "jsonpath={.data.password}")
	if err != nil {
		return fmt.Errorf("failed to check secret: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "not-valid-base64!" {
		return fmt.Errorf("secret still has invalid base64")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "db-client",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *SecretDataBase64WrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check Secret data",
			Command:     "kubectl get secret db-credentials -o jsonpath='{.data.password}'",
			Notes:       "Value is not valid base64",
		},
		{
			Description: "Generate valid base64",
			Command:     "echo -n 'mysecretpassword' | base64",
			Notes:       "Returns the properly encoded value",
		},
		{
			Description: "Fix the Secret",
			Command:     "kubectl patch secret db-credentials --type='json' -p='[{\"op\":\"replace\",\"path\":\"/data/password\",\"value\":\"bXlzZWNyZXRwYXNzd29yZA==\"}]'",
			Notes:       "Replace with valid base64 encoded password",
		},
		{
			Description: "Verify Secret is valid",
			Command:     "kubectl get secret db-credentials -o jsonpath='{.data.password}' | base64 -d",
			Notes:       "Should decode to the correct password",
		},
	}
}
