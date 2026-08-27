package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSPodSecurityPolicyLab{})
}

type CKSPodSecurityPolicyLab struct {
	BaseLab
}

func (l *CKSPodSecurityPolicyLab) ID() string             { return "cks_pod_security_policy" }
func (l *CKSPodSecurityPolicyLab) Title() string          { return "Configure Pod Security Policy" }
func (l *CKSPodSecurityPolicyLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSPodSecurityPolicyLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSPodSecurityPolicyLab) EstimatedTime() int     { return 25 }
func (l *CKSPodSecurityPolicyLab) Cert() Cert             { return CertCKS }
func (l *CKSPodSecurityPolicyLab) DomainWeight() int      { return 20 }
func (l *CKSPodSecurityPolicyLab) Tags() []string {
	return []string{"cks", "pod-security-policy", "admission", "microservice-vulns"}
}

func (l *CKSPodSecurityPolicyLab) Description() string {
	return `The cluster does not have Pod Security Policies (PSP) enabled. Pods can run
with privileged containers, use host namespaces, and have full capabilities.

Your task: Create a Pod Security Policy 'restricted-psp' that:
- Disallows privileged containers
- Requires runAsNonRoot
- Drops all capabilities
- Is restricted to the 'restricted-ns' namespace via a RoleBinding`
}

func (l *CKSPodSecurityPolicyLab) Hints() []string {
	return []string{
		"Create a PSP with privileged: false",
		"Set requiredDropCapabilities to ALL",
		"Create a ClusterRole and RoleBinding for the PSP",
	}
}

func (l *CKSPodSecurityPolicyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSPodSecurityPolicyLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSPodSecurityPolicyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "psp", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get PSPs: %w", err)
	}
	if strings.Contains(output, "restricted-psp") {
		return nil
	}
	return fmt.Errorf("restricted-psp not found")
}

func (l *CKSPodSecurityPolicyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create Pod Security Policy", Command: `cat <<EOF | kubectl apply -f -
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: restricted-psp
spec:
  privileged: false
  runAsUser:
    rule: MustRunAsNonRoot
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
  requiredDropCapabilities:
  - ALL
  volumes:
  - configMap
  - emptyDir
  - secret
  hostNetwork: false
  hostIPC: false
  hostPID: false
EOF`},
		{Description: "Create ClusterRole for PSP", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: restricted-psp-user
rules:
- apiGroups: ["policy"]
  resourceNames: ["restricted-psp"]
  resources: ["podsecuritypolicies"]
  verbs: ["use"]
EOF`},
		{Description: "Create RoleBinding", Command: `cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: restricted-psp-binding
  namespace: restricted-ns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: restricted-psp-user
subjects:
- kind: Group
  name: system:authenticated
  apiGroup: rbac.authorization.k8s.io
EOF`},
	}
}
