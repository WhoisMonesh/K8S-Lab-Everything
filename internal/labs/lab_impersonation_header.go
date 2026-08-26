package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ImpersonationHeader{})
}

type ImpersonationHeader struct {
	BaseLab
}

func (l *ImpersonationHeader) ID() string            { return "impersonation_header" }
func (l *ImpersonationHeader) Title() string         { return "Impersonation Header Denied" }
func (l *ImpersonationHeader) Category() Category    { return CategoryRBAC }
func (l *ImpersonationHeader) Difficulty() Difficulty { return DifficultyHard }
func (l *ImpersonationHeader) EstimatedTime() int    { return 20 }
func (l *ImpersonationHeader) Tags() []string        { return []string{"rbac", "impersonation", "security"} }

func (l *ImpersonationHeader) Description() string {
	return `A user is trying to impersonate another user but lacks the necessary permissions.
Fix the RBAC configuration to allow impersonation.`
}

func (l *ImpersonationHeader) Hints() []string {
	return []string{
		"Check if user has impersonation permissions",
		"Look at the impersonate verb in RBAC",
		"Create appropriate ClusterRole and binding",
	}
}

func (l *ImpersonationHeader) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ImpersonationHeader) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: impersonator
rules:
- apiGroups: [""]
  resources: ["users"]
  verbs: ["impersonate"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: impersonator-binding
subjects:
- kind: User
  name: admin
roleRef:
  kind: ClusterRole
  name: impersonator
  apiGroup: rbac.authorization.k8s.io`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ImpersonationHeader) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "auth", "can-i", "impersonate", "users/other-user",
		"--as=admin")
	if err != nil {
		return err
	}
	if output != "yes" {
		return fmt.Errorf("impersonation not allowed")
	}
	return nil
}

func (l *ImpersonationHeader) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check impersonation permissions", Command: "kubectl auth can-i impersonate users/other-user --as=admin"},
		{Description: "Fix ClusterRole", Command: "kubectl patch clusterrole impersonator --type='json' -p='[{\"op\": \"add\", \"path\": \"/rules/0/verbs/-\", \"value\": \"impersonate\"}]'"},
	}
}
