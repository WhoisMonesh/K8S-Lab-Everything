package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&RBACPermissionDeniedLab{})
}

type RBACPermissionDeniedLab struct {
	BaseLab
}

func (l *RBACPermissionDeniedLab) ID() string {
	return "rbac_permission_denied"
}

func (l *RBACPermissionDeniedLab) Title() string {
	return "RBAC Permission Denied"
}

func (l *RBACPermissionDeniedLab) Category() Category {
	return CategoryRBAC
}

func (l *RBACPermissionDeniedLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *RBACPermissionDeniedLab) Description() string {
	return `A developer user 'john' cannot create pods in the 'development' namespace.
The user is getting "forbidden" errors when trying to create resources.

Your task: Fix the RBAC configuration to allow john to create pods in the development namespace.`
}

func (l *RBACPermissionDeniedLab) Hints() []string {
	return []string{
		"Check if the user has a Role or ClusterRole assigned",
		"Look for RoleBindings in the development namespace",
		"The Role might be missing required permissions (verbs)",
		"Use kubectl auth can-i to test permissions",
	}
}

func (l *RBACPermissionDeniedLab) EstimatedTime() int {
	return 20
}

func (l *RBACPermissionDeniedLab) Tags() []string {
	return []string{"rbac", "roles", "rolebindings", "permissions", "security"}
}

func (l *RBACPermissionDeniedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *RBACPermissionDeniedLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create the development namespace
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: development
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	// Create a Role with incomplete permissions (missing 'create' verb)
	role := `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: developer-role
  namespace: development
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]
`
	if err := kubectlApply(ctx, kubeconfigPath, role); err != nil {
		return fmt.Errorf("creating role: %w", err)
	}

	// Create a RoleBinding for user john
	roleBinding := `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: developer-binding
  namespace: development
subjects:
- kind: User
  name: john
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: developer-role
  apiGroup: rbac.authorization.k8s.io
`
	if err := kubectlApply(ctx, kubeconfigPath, roleBinding); err != nil {
		return fmt.Errorf("creating rolebinding: %w", err)
	}

	return nil
}

func (l *RBACPermissionDeniedLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(5 * time.Second)
	// Test if john can create pods (should fail)
	_, err := kubectl(ctx, kubeconfigPath, "auth", "can-i", "create", "pods",
		"--namespace=development", "--as=john")
	// We expect this to return "no", but kubectl auth can-i returns exit code 1 for "no"
	_ = err // Ignore the error, it's expected
	return nil
}

func (l *RBACPermissionDeniedLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Test if john can create pods (should succeed)
	output, err := kubectl(ctx, kubeconfigPath, "auth", "can-i", "create", "pods",
		"--namespace=development", "--as=john")
	if err != nil {
		return fmt.Errorf("user 'john' still cannot create pods: %w", err)
	}

	// kubectl auth can-i returns "yes" or "no"
	output = strings.TrimSpace(output)
	if output != "yes" {
		return fmt.Errorf("user 'john' still cannot create pods (got: %s)", output)
	}

	// Also verify the role exists and has the correct permissions
	roleOutput, err := kubectl(ctx, kubeconfigPath, "get", "role", "developer-role",
		"-n", "development", "-o", "jsonpath={.rules[0].verbs}")
	if err != nil {
		return fmt.Errorf("failed to check role: %w", err)
	}

	if !strings.Contains(roleOutput, "create") {
		return fmt.Errorf("role still missing 'create' permission")
	}

	return nil
}

func (l *RBACPermissionDeniedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check if user can create pods",
			Command:     "kubectl auth can-i create pods --namespace=development --as=john",
			Notes:       "This should return 'no', confirming the permission issue",
		},
		{
			Description: "List roles in the development namespace",
			Command:     "kubectl get roles -n development",
			Notes:       "You should see a 'developer-role'",
		},
		{
			Description: "Examine the developer role",
			Command:     "kubectl describe role developer-role -n development",
			Notes:       "Notice the verbs are only 'get', 'list', 'watch' - missing 'create'",
		},
		{
			Description: "Check the rolebinding",
			Command:     "kubectl get rolebinding developer-binding -n development -o yaml",
			Notes:       "Verify it binds user 'john' to 'developer-role'",
		},
		{
			Description: "Fix the Role to include create permission",
			Command:     "kubectl edit role developer-role -n development",
			Notes:       "Add 'create' to the verbs list: verbs: ['get', 'list', 'watch', 'create']",
		},
		{
			Description: "Verify the fix",
			Command:     "kubectl auth can-i create pods --namespace=development --as=john",
			Notes:       "This should now return 'yes'",
		},
		{
			Description: "Test creating a pod as john (optional)",
			Command:     "kubectl run test-pod --image=nginx -n development --as=john",
			Notes:       "This should succeed now",
		},
	}
}
