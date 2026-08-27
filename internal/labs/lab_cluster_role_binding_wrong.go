package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ClusterRoleBindingWrong{})
}

type ClusterRoleBindingWrong struct {
	BaseLab
}

func (l *ClusterRoleBindingWrong) ID() string             { return "cluster_role_binding_wrong" }
func (l *ClusterRoleBindingWrong) Title() string          { return "ClusterRoleBinding References Wrong Role" }
func (l *ClusterRoleBindingWrong) Category() Category     { return CategoryRBAC }
func (l *ClusterRoleBindingWrong) Difficulty() Difficulty { return DifficultyMedium }
func (l *ClusterRoleBindingWrong) EstimatedTime() int     { return 15 }
func (l *ClusterRoleBindingWrong) Tags() []string {
	return []string{"rbac", "clusterrolebinding", "permissions"}
}

func (l *ClusterRoleBindingWrong) Description() string {
	return `A ClusterRoleBinding is referencing a ClusterRole that doesn't exist.
A service account cannot perform its required operations. Fix the binding.`
}

func (l *ClusterRoleBindingWrong) Hints() []string {
	return []string{
		"Check the ClusterRoleBinding configuration",
		"Verify the ClusterRole exists",
		"Fix the roleRef to point to an existing ClusterRole",
	}
}

func (l *ClusterRoleBindingWrong) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ClusterRoleBindingWrong) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pod-reader
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: read-pods-binding
subjects:
- kind: ServiceAccount
  name: default
  namespace: default
roleRef:
  kind: ClusterRole
  name: pod-reader-nonexistent
  apiGroup: rbac.authorization.k8s.io`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ClusterRoleBindingWrong) Verify(ctx context.Context, kubeconfigPath string) error {
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

func (l *ClusterRoleBindingWrong) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check ClusterRoleBinding", Command: "kubectl describe clusterrolebinding read-pods-binding"},
		{Description: "List ClusterRoles", Command: "kubectl get clusterroles | grep pod"},
		{Description: "Fix binding", Command: "kubectl patch clusterrolebinding read-pods-binding -p '{\"roleRef\":{\"name\":\"pod-reader\"}}'"},
	}
}
