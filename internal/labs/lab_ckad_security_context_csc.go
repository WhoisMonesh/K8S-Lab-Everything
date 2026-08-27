package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADSecurityContextCSCLab{})
}

type CKADSecurityContextCSCLab struct {
	BaseLab
}

func (l *CKADSecurityContextCSCLab) ID() string {
	return "ckad_security_context_csc"
}

func (l *CKADSecurityContextCSCLab) Title() string {
	return "Configure Container SecurityContext"
}

func (l *CKADSecurityContextCSCLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADSecurityContextCSCLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADSecurityContextCSCLab) Cert() Cert             { return CertCKAD }
func (l *CKADSecurityContextCSCLab) DomainWeight() int      { return 25 }
func (l *CKADSecurityContextCSCLab) EstimatedTime() int     { return 15 }
func (l *CKADSecurityContextCSCLab) Tags() []string {
	return []string{"security-context", "container", "privilege"}
}

func (l *CKADSecurityContextCSCLab) Description() string {
	return `A container needs specific security settings. Configure the Container
SecurityContext to disable privilege escalation and run as non-root.

Your task: Add the Container SecurityContext configuration.`
}

func (l *CKADSecurityContextCSCLab) Hints() []string {
	return []string{
		"Use securityContext at the container level",
		"Set allowPrivilegeEscalation: false",
		"Set runAsNonRoot: true",
	}
}

func (l *CKADSecurityContextCSCLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADSecurityContextCSCLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: secure-container
  labels:
    app: secure-container
spec:
  containers:
  - name: app
    image: nginx:alpine
    ports:
    - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADSecurityContextCSCLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "secure-container",
		"-o", "jsonpath={.spec.containers[0].securityContext.allowPrivilegeEscalation}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) != "false" {
		return fmt.Errorf("allowPrivilegeEscalation not set to false (current: %s)", output)
	}
	return nil
}

func (l *CKADSecurityContextCSCLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod", Command: "kubectl edit pod secure-container"},
		{Description: "Add container securityContext", Command: "Add securityContext with allowPrivilegeEscalation: false and runAsNonRoot: true"},
		{Description: "Verify", Command: "kubectl get pod secure-container -o yaml | grep -A 3 securityContext"},
	}
}
