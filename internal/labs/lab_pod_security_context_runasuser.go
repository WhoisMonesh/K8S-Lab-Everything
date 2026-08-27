package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodSecurityContextRunAsUserLab{})
}

type PodSecurityContextRunAsUserLab struct {
	BaseLab
}

func (l *PodSecurityContextRunAsUserLab) ID() string {
	return "pod_security_context_runasuser"
}

func (l *PodSecurityContextRunAsUserLab) Title() string {
	return "runAsUser Violation"
}

func (l *PodSecurityContextRunAsUserLab) Category() Category {
	return CategorySecurity
}

func (l *PodSecurityContextRunAsUserLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodSecurityContextRunAsUserLab) Description() string {
	return `A PodSecurityPolicy (or Pod Security Admission) blocks pods running
as root (UID 0). A pod 'restricted-pod' is configured with runAsUser: 0
which violates the policy.

Your task: Fix the security context to run as a non-root user.`
}

func (l *PodSecurityContextRunAsUserLab) Hints() []string {
	return []string{
		"Check the pod security context",
		"runAsUser: 0 means root which may be restricted",
		"Set runAsUser to a non-zero value like 1000",
	}
}

func (l *PodSecurityContextRunAsUserLab) EstimatedTime() int {
	return 15
}

func (l *PodSecurityContextRunAsUserLab) Tags() []string {
	return []string{"security", "runasuser", "pod-security", "nonroot"}
}

func (l *PodSecurityContextRunAsUserLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodSecurityContextRunAsUserLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: restricted-pod
  namespace: default
  labels:
    app: restricted-pod
spec:
  securityContext:
    runAsUser: 0
    runAsNonRoot: false
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'whoami && sleep 3600']
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodSecurityContextRunAsUserLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodSecurityContextRunAsUserLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "restricted-pod",
		"-o", "jsonpath={.spec.securityContext.runAsUser}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "0" {
		return fmt.Errorf("runAsUser is still 0 (root)")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "restricted-pod",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check status: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *PodSecurityContextRunAsUserLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check security context",
			Command:     "kubectl get pod restricted-pod -o yaml | grep -A 3 securityContext",
			Notes:       "runAsUser is 0 and runAsNonRoot is false",
		},
		{
			Description: "Fix security context",
			Command:     "kubectl edit pod restricted-pod",
			Notes:       "Set runAsUser to 1000 and runAsNonRoot to true",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod restricted-pod",
			Notes:       "Pod should now be Running",
		},
	}
}
