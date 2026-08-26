package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ServiceAccountMissingPermissions{})
}

type ServiceAccountMissingPermissions struct {
	BaseLab
}

func (l *ServiceAccountMissingPermissions) ID() string { return "service_account_missing_permissions" }
func (l *ServiceAccountMissingPermissions) Title() string {
	return "Service Account Missing Permissions"
}
func (l *ServiceAccountMissingPermissions) Category() Category     { return CategoryRBAC }
func (l *ServiceAccountMissingPermissions) Difficulty() Difficulty { return DifficultyMedium }
func (l *ServiceAccountMissingPermissions) EstimatedTime() int     { return 15 }
func (l *ServiceAccountMissingPermissions) Tags() []string {
	return []string{"rbac", "serviceaccount", "permissions"}
}

func (l *ServiceAccountMissingPermissions) Description() string {
	return `A pod using a service account cannot perform required API operations.
The service account is missing the necessary RBAC permissions. Grant them.`
}

func (l *ServiceAccountMissingPermissions) Hints() []string {
	return []string{
		"Check the service account's RoleBindings",
		"Verify the Role has required permissions",
		"Create a RoleBinding if missing",
	}
}

func (l *ServiceAccountMissingPermissions) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceAccountMissingPermissions) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: lister
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pod-lister
  namespace: default
rules:
- apiGroups: [""]
  resources: ["services"]
  verbs: ["get", "list"]`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ServiceAccountMissingPermissions) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "auth", "can-i", "list", "pods",
		"--as=system:serviceaccount:default:lister")
	if err != nil {
		return err
	}
	if output != "yes" {
		return fmt.Errorf("service account cannot list pods")
	}
	return nil
}

func (l *ServiceAccountMissingPermissions) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check role", Command: "kubectl describe role pod-lister"},
		{Description: "Fix role resources", Command: "kubectl patch role pod-lister --type='json' -p='[{\"op\": \"replace\", \"path\": \"/rules/0/resources/0\", \"value\": \"pods\"}]'"},
		{Description: "Create RoleBinding", Command: "kubectl create rolebinding lister-binding --role=pod-lister --serviceaccount=default:lister"},
	}
}
