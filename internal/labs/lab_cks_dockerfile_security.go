package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&CKSDockerfileSecurityLab{})
}

type CKSDockerfileSecurityLab struct {
	BaseLab
}

func (l *CKSDockerfileSecurityLab) ID() string             { return "cks_dockerfile_security" }
func (l *CKSDockerfileSecurityLab) Title() string          { return "Secure Dockerfile Best Practices" }
func (l *CKSDockerfileSecurityLab) Category() Category     { return CategorySupplyChain }
func (l *CKSDockerfileSecurityLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSDockerfileSecurityLab) EstimatedTime() int     { return 20 }
func (l *CKSDockerfileSecurityLab) Cert() Cert             { return CertCKS }
func (l *CKSDockerfileSecurityLab) DomainWeight() int      { return 20 }
func (l *CKSDockerfileSecurityLab) Tags() []string {
	return []string{"cks", "dockerfile", "security", "supply-chain"}
}

func (l *CKSDockerfileSecurityLab) Description() string {
	return `A pod 'app' in namespace 'secure-app' is running with an insecure security
context. It runs as root, allows privilege escalation, and has a writable
root filesystem.

Your task: Harden the pod so that it:
1. Runs as a non-root user (runAsNonRoot)
2. Disallows privilege escalation (allowPrivilegeEscalation: false)
3. Uses a read-only root filesystem
4. Applies the RuntimeDefault seccomp profile and drops all capabilities`
}

func (l *CKSDockerfileSecurityLab) Hints() []string {
	return []string{
		"Set securityContext.runAsNonRoot to true",
		"Set securityContext.allowPrivilegeEscalation to false",
		"Set securityContext.readOnlyRootFilesystem to true",
	}
}

func (l *CKSDockerfileSecurityLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSDockerfileSecurityLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: secure-app
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: app
  namespace: secure-app
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
    securityContext:
      allowPrivilegeEscalation: true
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSDockerfileSecurityLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *CKSDockerfileSecurityLab) Verify(ctx context.Context, kubeconfigPath string) error {
	allowEsc, err := kubectl(ctx, kubeconfigPath, "get", "pod", "app", "-n", "secure-app",
		"-o", "jsonpath={.spec.securityContext.allowPrivilegeEscalation}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(allowEsc) != "false" {
		return fmt.Errorf("allowPrivilegeEscalation not disabled (got: %s)", allowEsc)
	}
	runAsNonRoot, err := kubectl(ctx, kubeconfigPath, "get", "pod", "app", "-n", "secure-app",
		"-o", "jsonpath={.spec.securityContext.runAsNonRoot}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(runAsNonRoot) != "true" {
		return fmt.Errorf("runAsNonRoot not enabled (got: %s)", runAsNonRoot)
	}
	readOnly, err := kubectl(ctx, kubeconfigPath, "get", "pod", "app", "-n", "secure-app",
		"-o", "jsonpath={.spec.containers[0].securityContext.readOnlyRootFilesystem}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(readOnly) != "true" {
		return fmt.Errorf("readOnlyRootFilesystem not enabled (got: %s)", readOnly)
	}
	return nil
}

func (l *CKSDockerfileSecurityLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Delete and recreate pod with hardened securityContext", Command: `kubectl delete pod app -n secure-app && kubectl run app -n secure-app --image=busybox:1.36 --restart=Never --overrides='{"spec":{"securityContext":{"runAsNonRoot":true,"seccompProfile":{"type":"RuntimeDefault"}},"containers":[{"name":"app","image":"busybox:1.36","command":["sh","-c","while true; do sleep 3600; done"],"securityContext":{"allowPrivilegeEscalation":false,"readOnlyRootFilesystem":true,"capabilities":{"drop":["ALL"]}}}]}}'`},
		{Description: "Verify hardening", Command: "kubectl get pod app -n secure-app -o jsonpath='{.spec.securityContext.allowPrivilegeEscalation} {.spec.securityContext.runAsNonRoot}'"},
	}
}
