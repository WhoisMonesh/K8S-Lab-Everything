package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADSecurityContextPSCLab{})
}

type CKADSecurityContextPSCLab struct {
	BaseLab
}

func (l *CKADSecurityContextPSCLab) ID() string {
	return "ckad_security_context_psc"
}

func (l *CKADSecurityContextPSCLab) Title() string {
	return "Configure Pod SecurityContext"
}

func (l *CKADSecurityContextPSCLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADSecurityContextPSCLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADSecurityContextPSCLab) Cert() Cert             { return CertCKAD }
func (l *CKADSecurityContextPSCLab) DomainWeight() int      { return 25 }
func (l *CKADSecurityContextPSCLab) EstimatedTime() int     { return 15 }
func (l *CKADSecurityContextPSCLab) Tags() []string {
	return []string{"security-context", "pod", "run-as-user"}
}

func (l *CKADSecurityContextPSCLab) Description() string {
	return `A pod needs to run as a specific user for security compliance. Configure
the Pod SecurityContext to run as user 1000 and group 3000.

Your task: Add the Pod SecurityContext configuration.`
}

func (l *CKADSecurityContextPSCLab) Hints() []string {
	return []string{
		"Use securityContext at the pod level",
		"Set runAsUser and runAsGroup",
		"fsGroup can be set for volume ownership",
	}
}

func (l *CKADSecurityContextPSCLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADSecurityContextPSCLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: secure-app
  labels:
    app: secure-app
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'id && sleep 3600']`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADSecurityContextPSCLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "secure-app",
		"-o", "jsonpath={.spec.securityContext.runAsUser}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) != "1000" {
		return fmt.Errorf("runAsUser not set to 1000 (current: %s)", output)
	}
	return nil
}

func (l *CKADSecurityContextPSCLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod", Command: "kubectl edit pod secure-app"},
		{Description: "Add securityContext", Command: "Add securityContext with runAsUser: 1000 and runAsGroup: 3000"},
		{Description: "Verify", Command: "kubectl exec secure-app -- id"},
	}
}
