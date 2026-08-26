package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodSecurityPolicyViolation{})
}

type PodSecurityPolicyViolation struct {
	BaseLab
}

func (l *PodSecurityPolicyViolation) ID() string            { return "pod_security_policy_violation" }
func (l *PodSecurityPolicyViolation) Title() string         { return "Pod Security Policy Violation" }
func (l *PodSecurityPolicyViolation) Category() Category    { return CategorySecurity }
func (l *PodSecurityPolicyViolation) Difficulty() Difficulty { return DifficultyMedium }
func (l *PodSecurityPolicyViolation) EstimatedTime() int    { return 15 }
func (l *PodSecurityPolicyViolation) Tags() []string        { return []string{"security", "psp", "admission"} }

func (l *PodSecurityPolicyViolation) Description() string {
	return `A pod is being rejected by the PodSecurityPolicy admission controller.
The pod violates the security policy. Fix the pod spec to comply with the policy.`
}

func (l *PodSecurityPolicyViolation) Hints() []string {
	return []string{
		"Check the PodSecurityPolicy",
		"Look at the pod securityContext",
		"Ensure runAsNonRoot is set to true",
	}
}

func (l *PodSecurityPolicyViolation) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodSecurityPolicyViolation) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: restricted
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  runAsUser:
    rule: MustRunAsNonRoot
  seLinux:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  volumes:
    - '*'
---
apiVersion: v1
kind: Pod
metadata:
  name: restricted-pod
spec:
  securityContext:
    runAsUser: 0
  containers:
  - name: nginx
    image: nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PodSecurityPolicyViolation) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "restricted-pod",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("pod not running: %s", output)
	}
	return nil
}

func (l *PodSecurityPolicyViolation) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check PodSecurityPolicy", Command: "kubectl describe psp restricted"},
		{Description: "Fix pod securityContext", Command: "kubectl edit pod restricted-pod"},
		{Description: "Set runAsNonRoot", Command: "Add securityContext: {runAsNonRoot: true, runAsUser: 1000}"},
	}
}
