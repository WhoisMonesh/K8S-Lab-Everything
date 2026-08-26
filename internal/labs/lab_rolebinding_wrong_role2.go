package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&RoleBindingWrongRole{})
}

type RoleBindingWrongRole struct {
	BaseLab
}

func (l *RoleBindingWrongRole) ID() string            { return "rolebinding_wrong_role2" }
func (l *RoleBindingWrongRole) Title() string         { return "RoleBinding References Wrong Role" }
func (l *RoleBindingWrongRole) Category() Category    { return CategoryRBAC }
func (l *RoleBindingWrongRole) Difficulty() Difficulty { return DifficultyMedium }
func (l *RoleBindingWrongRole) EstimatedTime() int    { return 15 }
func (l *RoleBindingWrongRole) Tags() []string        { return []string{"rbac", "rolebinding", "permissions"} }

func (l *RoleBindingWrongRole) Description() string {
	return `A RoleBinding references a Role that doesn't exist.
A service account cannot perform its required operations. Fix the binding.`
}

func (l *RoleBindingWrongRole) Hints() []string {
	return []string{
		"Check the RoleBinding configuration",
		"Verify the Role exists",
		"Fix the roleRef to point to an existing Role",
	}
}

func (l *RoleBindingWrongRole) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *RoleBindingWrongRole) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-reader
  namespace: default
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: read-pods-binding
  namespace: default
subjects:
- kind: ServiceAccount
  name: default
  namespace: default
roleRef:
  kind: Role
  name: pod-reader-nonexistent
  apiGroup: rbac.authorization.k8s.io`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *RoleBindingWrongRole) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "auth", "can-i", "get", "pods",
		"--as=system:serviceaccount:default:default")
	if err != nil {
		return err
	}
	if output != "yes" {
		return fmt.Errorf("service account cannot get pods")
	}
	return nil
}

func (l *RoleBindingWrongRole) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check RoleBinding", Command: "kubectl describe rolebinding read-pods-binding"},
		{Description: "List Roles", Command: "kubectl get roles"},
		{Description: "Fix binding", Command: "kubectl patch rolebinding read-pods-binding -p '{\"roleRef\":{\"name\":\"pod-reader\"}}'"},
	}
}
