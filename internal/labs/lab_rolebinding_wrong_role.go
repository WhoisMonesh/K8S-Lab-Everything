package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&RoleBindingWrongRoleLab{}) }

type RoleBindingWrongRoleLab struct{ BaseLab }

func (l *RoleBindingWrongRoleLab) ID() string          { return "rolebinding_wrong_role" }
func (l *RoleBindingWrongRoleLab) Title() string        { return "RoleBinding References Missing Role" }
func (l *RoleBindingWrongRoleLab) Category() Category   { return CategoryRBAC }
func (l *RoleBindingWrongRoleLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *RoleBindingWrongRoleLab) EstimatedTime() int   { return 15 }
func (l *RoleBindingWrongRoleLab) Tags() []string {
	return []string{"rbac", "rolebinding", "role", "permissions"}
}
func (l *RoleBindingWrongRoleLab) Description() string {
	return `A ServiceAccount 'app-sa' in namespace 'secure-ns' should have
read-only access to pods, but kubectl auth can-i get pods --as=system:serviceaccount:secure-ns:app-sa
returns 'no'.

The RoleBinding 'pod-reader-binding' exists but references a Role
named 'pod-editor-role' (which grants edit access but doesn't exist).

Your task: Create the RoleBinding that correctly grants read-only
(pods get/watch/list) access to app-sa via a properly defined Role.`
}
func (l *RoleBindingWrongRoleLab) Hints() []string {
	return []string{
		"Check: kubectl describe rolebinding pod-reader-binding -n secure-ns",
		"It references 'pod-editor-role' which doesn't exist",
		"Create Role 'pod-reader' with get, list, watch on pods",
		"Patch the RoleBinding to reference the correct Role",
	}
}

func (l *RoleBindingWrongRoleLab) Break(ctx context.Context, kp string) error {
	if _, err := kubectl(ctx, kp, "create", "ns", "secure-ns"); err != nil {
		return err
	}
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: secure-ns
`
	sa := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: app-sa
  namespace: secure-ns
`
	rb := `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: pod-reader-binding
  namespace: secure-ns
subjects:
- kind: ServiceAccount
  name: app-sa
  namespace: secure-ns
roleRef:
  kind: Role
  name: pod-editor-role
  apiGroup: rbac.authorization.k8s.io
`
	kubectlApply(ctx, kp, ns)
	kubectlApply(ctx, kp, sa)
	return kubectlApply(ctx, kp, rb)
}

func (l *RoleBindingWrongRoleLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(3 * time.Second)
	return nil
}

func (l *RoleBindingWrongRoleLab) Verify(ctx context.Context, kp string) error {
	// Check role exists
	role, err := kubectl(ctx, kp, "get", "role", "pod-reader", "-n", "secure-ns", "-o", "jsonpath={.metadata.name}")
	if err != nil || !strings.Contains(role, "pod-reader") {
		return fmt.Errorf("Role 'pod-reader' not found in secure-ns")
	}
	// Check binding references correct role
	bindingRole, _ := kubectl(ctx, kp, "get", "rolebinding", "pod-reader-binding", "-n", "secure-ns",
		"-o", "jsonpath={.roleRef.name}")
	if bindingRole != "pod-reader" {
		return fmt.Errorf("RoleBinding still references wrong role: %s", bindingRole)
	}
	// Check auth
	auth, _ := kubectl(ctx, kp, "auth", "can-i", "get", "pods",
		"--as=system:serviceaccount:secure-ns:app-sa", "-n", "secure-ns")
	if auth != "yes" {
		return fmt.Errorf("app-sa still cannot get pods (auth: %s)", auth)
	}
	return nil
}

func (l *RoleBindingWrongRoleLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Diagnose the binding", Command: "kubectl describe rolebinding pod-reader-binding -n secure-ns", Notes: "roleRef is pod-editor-role — which doesn't exist"},
		{Description: "Create the correct Role", Command: `kubectl create role pod-reader --verb=get,list,watch --resource=pods -n secure-ns`, Notes: "Defines read-only permissions on pods"},
		{Description: "Patch the binding", Command: `kubectl patch rolebinding pod-reader-binding -n secure-ns -p '{"roleRef":{"name":"pod-reader"}}'`, Notes: "Redirects binding to the correct Role"},
		{Description: "Verify", Command: "kubectl auth can-i get pods --as=system:serviceaccount:secure-ns:app-sa -n secure-ns", Notes: "Returns 'yes'"},
	}
}
