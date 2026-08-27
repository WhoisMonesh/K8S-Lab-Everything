package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodSeccompRuntimeDefaultLab{})
}

type PodSeccompRuntimeDefaultLab struct {
	BaseLab
}

func (l *PodSeccompRuntimeDefaultLab) ID() string {
	return "pod_seccomp_runtime_default"
}

func (l *PodSeccompRuntimeDefaultLab) Title() string {
	return "Seccomp Profile Runtime Default"
}

func (l *PodSeccompRuntimeDefaultLab) Category() Category {
	return CategorySecurity
}

func (l *PodSeccompRuntimeDefaultLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodSeccompRuntimeDefaultLab) Description() string {
	return `A pod 'seccomp-app' has a seccomp profile set to Unconfined which
provides no syscall filtering. The security policy requires using the
runtime default profile.

Your task: Fix the seccomp profile to use RuntimeDefault.`
}

func (l *PodSeccompRuntimeDefaultLab) Hints() []string {
	return []string{
		"Check the pod's security context",
		"Unconfined means no seccomp filtering",
		"Set seccompProfile.type to RuntimeDefault",
	}
}

func (l *PodSeccompRuntimeDefaultLab) EstimatedTime() int {
	return 10
}

func (l *PodSeccompRuntimeDefaultLab) Tags() []string {
	return []string{"security", "seccomp", "syscall", "hardening"}
}

func (l *PodSeccompRuntimeDefaultLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodSeccompRuntimeDefaultLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: seccomp-app
  namespace: default
spec:
  securityContext:
    seccompProfile:
      type: Unconfined
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'sleep 3600']
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodSeccompRuntimeDefaultLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodSeccompRuntimeDefaultLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "seccomp-app",
		"-o", "jsonpath={.spec.securityContext.seccompProfile.type}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) == "Unconfined" {
		return fmt.Errorf("seccomp profile is still Unconfined")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "seccomp-app",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check status: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *PodSeccompRuntimeDefaultLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check seccomp profile",
			Command:     "kubectl get pod seccomp-app -o yaml | grep -A 3 seccompProfile",
			Notes:       "type is Unconfined",
		},
		{
			Description: "Fix seccomp profile",
			Command:     "kubectl edit pod seccomp-app",
			Notes:       "Change type from Unconfined to RuntimeDefault",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod seccomp-app",
			Notes:       "Pod should now be Running with RuntimeDefault seccomp",
		},
	}
}
