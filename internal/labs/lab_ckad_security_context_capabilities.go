package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADSecurityContextCapabilitiesLab{})
}

type CKADSecurityContextCapabilitiesLab struct {
	BaseLab
}

func (l *CKADSecurityContextCapabilitiesLab) ID() string {
	return "ckad_security_context_capabilities"
}

func (l *CKADSecurityContextCapabilitiesLab) Title() string {
	return "Add/Drop Linux Capabilities"
}

func (l *CKADSecurityContextCapabilitiesLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADSecurityContextCapabilitiesLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKADSecurityContextCapabilitiesLab) Cert() Cert             { return CertCKAD }
func (l *CKADSecurityContextCapabilitiesLab) DomainWeight() int      { return 25 }
func (l *CKADSecurityContextCapabilitiesLab) EstimatedTime() int     { return 20 }
func (l *CKADSecurityContextCapabilitiesLab) Tags() []string {
	return []string{"capabilities", "linux", "security"}
}

func (l *CKADSecurityContextCapabilitiesLab) Description() string {
	return `A container needs specific Linux capabilities to perform network operations.
Add NET_ADMIN and NET_RAW capabilities while dropping all others.

Your task: Configure the container capabilities.`
}

func (l *CKADSecurityContextCapabilitiesLab) Hints() []string {
	return []string{
		"Use securityContext.capabilities.add and drop",
		"NET_ADMIN allows network administration",
		"NET_RAW allows raw socket access",
	}
}

func (l *CKADSecurityContextCapabilitiesLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADSecurityContextCapabilitiesLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: net-tools
  labels:
    app: net-tools
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'while true; do sleep 3600; done']`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADSecurityContextCapabilitiesLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "net-tools",
		"-o", "jsonpath={.spec.containers[0].securityContext.capabilities.add[*]}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if !strings.Contains(output, "NET_ADMIN") {
		return fmt.Errorf("NET_ADMIN capability not added")
	}
	return nil
}

func (l *CKADSecurityContextCapabilitiesLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod", Command: "kubectl edit pod net-tools"},
		{Description: "Add capabilities", Command: "Add securityContext.capabilities with add: [NET_ADMIN, NET_RAW] and drop: [ALL]"},
		{Description: "Verify", Command: "kubectl get pod net-tools -o yaml | grep -A 5 capabilities"},
	}
}
