package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&SeccompInvalidProfile{})
}

type SeccompInvalidProfile struct {
	BaseLab
}

func (l *SeccompInvalidProfile) ID() string             { return "seccomp_invalid_profile2" }
func (l *SeccompInvalidProfile) Title() string          { return "Pod Rejected - Invalid seccomp Profile" }
func (l *SeccompInvalidProfile) Category() Category     { return CategorySecurity }
func (l *SeccompInvalidProfile) Difficulty() Difficulty { return DifficultyHard }
func (l *SeccompInvalidProfile) EstimatedTime() int     { return 20 }
func (l *SeccompInvalidProfile) Tags() []string         { return []string{"security", "seccomp", "profiles"} }

func (l *SeccompInvalidProfile) Description() string {
	return `A pod is being rejected because it references a seccomp profile that doesn't exist.
Fix the seccomp profile configuration.`
}

func (l *SeccompInvalidProfile) Hints() []string {
	return []string{
		"Check the pod securityContext",
		"Look at the seccompProfile configuration",
		"Use RuntimeDefault or Unconfined profile",
	}
}

func (l *SeccompInvalidProfile) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SeccompInvalidProfile) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Pod
metadata:
  name: seccomp-pod
spec:
  securityContext:
    seccompProfile:
      type: Localhost
      localhostProfile: profiles/nonexistent.json
  containers:
  - name: nginx
    image: nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *SeccompInvalidProfile) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "seccomp-pod",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("pod not running: %s", output)
	}
	return nil
}

func (l *SeccompInvalidProfile) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod securityContext", Command: "kubectl get pod seccomp-pod -o jsonpath='{.spec.securityContext}'"},
		{Description: "Fix seccomp profile", Command: "kubectl edit pod seccomp-pod"},
		{Description: "Use RuntimeDefault", Command: "Change type to RuntimeDefault or Unconfined"},
	}
}
