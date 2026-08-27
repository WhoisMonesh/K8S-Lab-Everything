package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSContainerImmutabilityLab{})
}

type CKSContainerImmutabilityLab struct {
	BaseLab
}

func (l *CKSContainerImmutabilityLab) ID() string             { return "cks_container_immutability" }
func (l *CKSContainerImmutabilityLab) Title() string          { return "Enforce Container Immutability" }
func (l *CKSContainerImmutabilityLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSContainerImmutabilityLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSContainerImmutabilityLab) EstimatedTime() int     { return 20 }
func (l *CKSContainerImmutabilityLab) Cert() Cert             { return CertCKS }
func (l *CKSContainerImmutabilityLab) DomainWeight() int      { return 20 }
func (l *CKSContainerImmutabilityLab) Tags() []string {
	return []string{"cks", "immutability", "container", "security"}
}

func (l *CKSContainerImmutabilityLab) Description() string {
	return `Containers can be exec'd into and modified at runtime. This allows attackers
to inject malicious code or modify running applications.

Your task: Configure a security policy that prevents exec into pods in the
'readonly-apps' namespace by creating a RBAC Role that denies exec permissions.`
}

func (l *CKSContainerImmutabilityLab) Hints() []string {
	return []string{
		"Create a Role that denies exec and attach verbs",
		"Bind the Role to all authenticated users",
		"Test that kubectl exec is denied",
	}
}

func (l *CKSContainerImmutabilityLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSContainerImmutabilityLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: readonly-apps
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: secure-app
  namespace: readonly-apps
spec:
  containers:
  - name: app
    image: nginx:alpine
    ports:
    - containerPort: 80
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSContainerImmutabilityLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "role", "-n", "readonly-apps", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get roles: %w", err)
	}
	if strings.Contains(output, "exec") && strings.Contains(output, "deny") {
		return nil
	}
	return fmt.Errorf("immutability role not found")
}

func (l *CKSContainerImmutabilityLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create deny exec Role", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: deny-exec
  namespace: readonly-apps
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["pods/exec", "pods/attach"]
  verbs: ["create", "get", "list", "watch"]
EOF`},
		{Description: "Create RoleBinding", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: deny-exec-binding
  namespace: readonly-apps
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: deny-exec
subjects:
- kind: Group
  name: system:authenticated
  apiGroup: rbac.authorization.k8s.io
EOF`},
		{Description: "Verify", Command: "kubectl describe role deny-exec -n readonly-apps"},
	}
}
