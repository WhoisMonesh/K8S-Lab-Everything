package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSSeccompProfileLab{})
}

type CKSSeccompProfileLab struct {
	BaseLab
}

func (l *CKSSeccompProfileLab) ID() string             { return "cks_seccomp_profile" }
func (l *CKSSeccompProfileLab) Title() string          { return "Configure Seccomp Profile for Pods" }
func (l *CKSSeccompProfileLab) Category() Category     { return CategorySystemHardening }
func (l *CKSSeccompProfileLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSSeccompProfileLab) EstimatedTime() int     { return 20 }
func (l *CKSSeccompProfileLab) Cert() Cert             { return CertCKS }
func (l *CKSSeccompProfileLab) DomainWeight() int      { return 10 }
func (l *CKSSeccompProfileLab) Tags() []string {
	return []string{"cks", "seccomp", "pod-security", "system-hardening"}
}

func (l *CKSSeccompProfileLab) Description() string {
	return `A pod 'seccomp-app' in namespace 'hardened' runs without a seccomp profile.
This means the pod can make unrestricted system calls.

Your task: Configure the pod to use RuntimeDefault seccomp profile by adding
the appropriate securityContext settings.`
}

func (l *CKSSeccompProfileLab) Hints() []string {
	return []string{
		"Use securityContext.seccompProfile.type",
		"Set it to RuntimeDefault",
		"Delete and recreate the pod with the new security context",
	}
}

func (l *CKSSeccompProfileLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSSeccompProfileLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: hardened
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: seccomp-app
  namespace: hardened
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSSeccompProfileLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "seccomp-app", "-n", "hardened",
		"-o", "jsonpath={.spec.securityContext.seccompProfile.type}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "RuntimeDefault" {
		return nil
	}
	return fmt.Errorf("seccomp profile not set to RuntimeDefault (got: %s)", output)
}

func (l *CKSSeccompProfileLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Delete and recreate pod with seccomp", Command: `kubectl delete pod seccomp-app -n hardened && kubectl run seccomp-app -n hardened --image=busybox:1.36 --restart=Never --overrides='{"spec":{"securityContext":{"seccompProfile":{"type":"RuntimeDefault"}},"containers":[{"name":"seccomp-app","image":"busybox:1.36","command":["sh","-c","while true; do sleep 3600; done"]}]}}'`},
		{Description: "Verify seccomp profile", Command: "kubectl get pod seccomp-app -n hardened -o jsonpath='{.spec.securityContext.seccompProfile.type}'"},
	}
}
