package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSServiceAccountMinimizeLab{})
}

type CKSServiceAccountMinimizeLab struct {
	BaseLab
}

func (l *CKSServiceAccountMinimizeLab) ID() string             { return "cks_service_account_minimize" }
func (l *CKSServiceAccountMinimizeLab) Title() string          { return "Minimize ServiceAccount Permissions" }
func (l *CKSServiceAccountMinimizeLab) Category() Category     { return CategoryClusterHardening }
func (l *CKSServiceAccountMinimizeLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSServiceAccountMinimizeLab) EstimatedTime() int     { return 20 }
func (l *CKSServiceAccountMinimizeLab) Cert() Cert             { return CertCKS }
func (l *CKSServiceAccountMinimizeLab) DomainWeight() int      { return 15 }
func (l *CKSServiceAccountMinimizeLab) Tags() []string {
	return []string{"cks", "service-account", "rbac", "security"}
}

func (l *CKSServiceAccountMinimizeLab) Description() string {
	return `The 'monitor' ServiceAccount in the 'observability' namespace has been given
a ClusterRoleBinding to the 'cluster-admin' role, far exceeding its needs.

Your task: Replace the ClusterRoleBinding with a RoleBinding that grants only
get, list, and watch permissions on pods and services in the 'observability' namespace.`
}

func (l *CKSServiceAccountMinimizeLab) Hints() []string {
	return []string{
		"Delete the existing ClusterRoleBinding",
		"Create a Role with minimal permissions",
		"Create a RoleBinding linking the Role to the ServiceAccount",
	}
}

func (l *CKSServiceAccountMinimizeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSServiceAccountMinimizeLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: observability
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	sa := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: monitor
  namespace: observability
`
	if err := kubectlApply(ctx, kubeconfigPath, sa); err != nil {
		return fmt.Errorf("creating service account: %w", err)
	}

	crb := `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: monitor-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: monitor
  namespace: observability
`
	return kubectlApply(ctx, kubeconfigPath, crb)
}

func (l *CKSServiceAccountMinimizeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "clusterrolebindings", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get clusterrolebindings: %w", err)
	}
	if strings.Contains(output, "monitor-admin") {
		return fmt.Errorf("cluster-admin binding still exists")
	}

	rbOutput, err := kubectl(ctx, kubeconfigPath, "get", "rolebindings", "-n", "observability", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get rolebindings: %w", err)
	}
	if !strings.Contains(rbOutput, "monitor") {
		return fmt.Errorf("rolebinding for monitor not created")
	}
	return nil
}

func (l *CKSServiceAccountMinimizeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Delete cluster-admin binding", Command: "kubectl delete clusterrolebinding monitor-admin"},
		{Description: "Create minimal Role", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: monitor-reader
  namespace: observability
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list", "watch"]
EOF`},
		{Description: "Create RoleBinding", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: monitor-binding
  namespace: observability
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: monitor-reader
subjects:
- kind: ServiceAccount
  name: monitor
  namespace: observability
EOF`},
		{Description: "Verify", Command: "kubectl describe rolebinding monitor-binding -n observability"},
	}
}
