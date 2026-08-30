package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&CKSPrivilegeEscalationDetectLab{})
}

type CKSPrivilegeEscalationDetectLab struct {
	BaseLab
}

func (l *CKSPrivilegeEscalationDetectLab) ID() string { return "cks_privilege_escalation_detect" }
func (l *CKSPrivilegeEscalationDetectLab) Title() string {
	return "Detect Privilege Escalation Attempts"
}
func (l *CKSPrivilegeEscalationDetectLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSPrivilegeEscalationDetectLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSPrivilegeEscalationDetectLab) EstimatedTime() int     { return 30 }
func (l *CKSPrivilegeEscalationDetectLab) Cert() Cert             { return CertCKS }
func (l *CKSPrivilegeEscalationDetectLab) DomainWeight() int      { return 20 }
func (l *CKSPrivilegeEscalationDetectLab) Tags() []string {
	return []string{"cks", "privilege-escalation", "detection", "monitoring"}
}

func (l *CKSPrivilegeEscalationDetectLab) Description() string {
	return `A risky ClusterRoleBinding 'escalation-risk-binding' is present in the cluster.
It grants a ServiceAccount nearly cluster-admin level access ('*' on '*'),
which is a serious privilege escalation risk.

Your task: Auditing and removing this over-privileged binding so that the
risky access is revoked.`
}

func (l *CKSPrivilegeEscalationDetectLab) Hints() []string {
	return []string{
		"Describe the ClusterRoleBinding to understand what it grants",
		"Delete the over-privileged ClusterRoleBinding",
		"Optionally delete the broad ClusterRole too",
	}
}

func (l *CKSPrivilegeEscalationDetectLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSPrivilegeEscalationDetectLab) Break(ctx context.Context, kubeconfigPath string) error {
	sa := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: risky-sa
  namespace: default
`
	if err := kubectlApply(ctx, kubeconfigPath, sa); err != nil {
		return fmt.Errorf("creating serviceaccount: %w", err)
	}

	role := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: escalation-risk
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
`
	if err := kubectlApply(ctx, kubeconfigPath, role); err != nil {
		return fmt.Errorf("creating clusterrole: %w", err)
	}

	binding := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: escalation-risk-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: escalation-risk
subjects:
- kind: ServiceAccount
  name: risky-sa
  namespace: default
`
	return kubectlApply(ctx, kubeconfigPath, binding)
}

func (l *CKSPrivilegeEscalationDetectLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *CKSPrivilegeEscalationDetectLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "clusterrolebinding", "escalation-risk-binding", "-o", "name")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil
		}
		return fmt.Errorf("failed to check clusterrolebinding: %w", err)
	}
	if strings.TrimSpace(output) != "" {
		return fmt.Errorf("over-privileged clusterrolebinding still present")
	}
	return nil
}

func (l *CKSPrivilegeEscalationDetectLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Inspect the risky binding", Command: "kubectl describe clusterrolebinding escalation-risk-binding"},
		{Description: "Remove the over-privileged binding", Command: "kubectl delete clusterrolebinding escalation-risk-binding"},
		{Description: "Optionally delete the broad clusterrole", Command: "kubectl delete clusterrole escalation-risk"},
	}
}
