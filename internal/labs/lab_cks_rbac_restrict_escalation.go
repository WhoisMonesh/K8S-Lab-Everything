package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSRBACRestrictEscalationLab{})
}

type CKSRBACRestrictEscalationLab struct {
	BaseLab
}

func (l *CKSRBACRestrictEscalationLab) ID() string             { return "cks_rbac_restrict_escalation" }
func (l *CKSRBACRestrictEscalationLab) Title() string          { return "Prevent RBAC Privilege Escalation" }
func (l *CKSRBACRestrictEscalationLab) Category() Category     { return CategoryClusterHardening }
func (l *CKSRBACRestrictEscalationLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSRBACRestrictEscalationLab) EstimatedTime() int     { return 25 }
func (l *CKSRBACRestrictEscalationLab) Cert() Cert             { return CertCKS }
func (l *CKSRBACRestrictEscalationLab) DomainWeight() int      { return 15 }
func (l *CKSRBACRestrictEscalationLab) Tags() []string {
	return []string{"cks", "rbac", "escalation", "security"}
}

func (l *CKSRBACRestrictEscalationLab) Description() string {
	return `A Role 'pod-manager' in namespace 'apps' grants bind permission on
ClusterRoles, which allows privilege escalation by binding more powerful roles.

Your task: Remove the bind verb from the pod-manager Role and ensure it only
has permissions to manage pods.`
}

func (l *CKSRBACRestrictEscalationLab) Hints() []string {
	return []string{
		"Check the Role definition with kubectl describe role",
		"Remove the bind verb from the rules",
		"Ensure only pod management verbs remain",
	}
}

func (l *CKSRBACRestrictEscalationLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSRBACRestrictEscalationLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: apps
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	role := `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-manager
  namespace: apps
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["rbac.authorization.k8s.io"]
  resources: ["clusterroles"]
  verbs: ["bind"]
`
	return kubectlApply(ctx, kubeconfigPath, role)
}

func (l *CKSRBACRestrictEscalationLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "role", "pod-manager", "-n", "apps", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if strings.Contains(output, "bind") {
		return fmt.Errorf("bind verb still present in role")
	}
	if strings.Contains(output, "clusterroles") {
		return fmt.Errorf("clusterroles resource still in role rules")
	}
	return nil
}

func (l *CKSRBACRestrictEscalationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current Role rules", Command: "kubectl describe role pod-manager -n apps"},
		{Description: "Update Role to remove escalation", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-manager
  namespace: apps
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
EOF`},
		{Description: "Verify", Command: "kubectl describe role pod-manager -n apps"},
	}
}
