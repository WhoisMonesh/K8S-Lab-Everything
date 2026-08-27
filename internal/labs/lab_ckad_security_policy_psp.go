package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADSecurityPolicyPSPLab{})
}

type CKADSecurityPolicyPSPLab struct {
	BaseLab
}

func (l *CKADSecurityPolicyPSPLab) ID() string {
	return "ckad_security_policy_psp"
}

func (l *CKADSecurityPolicyPSPLab) Title() string {
	return "Configure Pod Security Standards"
}

func (l *CKADSecurityPolicyPSPLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADSecurityPolicyPSPLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKADSecurityPolicyPSPLab) Cert() Cert             { return CertCKAD }
func (l *CKADSecurityPolicyPSPLab) DomainWeight() int      { return 25 }
func (l *CKADSecurityPolicyPSPLab) EstimatedTime() int     { return 25 }
func (l *CKADSecurityPolicyPSPLab) Tags() []string {
	return []string{"pod-security", "standards", "enforcement"}
}

func (l *CKADSecurityPolicyPSPLab) Description() string {
	return `A namespace needs Pod Security Standards enforcement. Configure the
namespace to enforce the 'restricted' policy level.

Your task: Label the namespace to enforce Pod Security Standards.`
}

func (l *CKADSecurityPolicyPSPLab) Hints() []string {
	return []string{
		"Use pod-security.kubernetes.io/enforce label",
		"Set the enforcement level to 'restricted'",
		"Also set audit and warn labels for monitoring",
	}
}

func (l *CKADSecurityPolicyPSPLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADSecurityPolicyPSPLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: secure-ns`
	return kubectlApply(ctx, kubeconfigPath, ns)
}

func (l *CKADSecurityPolicyPSPLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "namespace", "secure-ns",
		"-o", "jsonpath={.metadata.labels}")
	if err != nil {
		return fmt.Errorf("failed to get namespace: %w", err)
	}
	if !strings.Contains(output, "restricted") {
		return fmt.Errorf("restricted enforcement not set")
	}
	return nil
}

func (l *CKADSecurityPolicyPSPLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Label namespace", Command: "kubectl label namespace secure-ns pod-security.kubernetes.io/enforce=restricted"},
		{Description: "Add audit label", Command: "kubectl label namespace secure-ns pod-security.kubernetes.io/audit=restricted"},
		{Description: "Add warn label", Command: "kubectl label namespace secure-ns pod-security.kubernetes.io/warn=restricted"},
		{Description: "Verify", Command: "kubectl get namespace secure-ns --show-labels"},
	}
}
