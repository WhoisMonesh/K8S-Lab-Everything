package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSPSSRestrictedLab{})
}

type CKSPSSRestrictedLab struct {
	BaseLab
}

func (l *CKSPSSRestrictedLab) ID() string             { return "cks_pss_restricted" }
func (l *CKSPSSRestrictedLab) Title() string          { return "Enforce Pod Security Standards Restricted" }
func (l *CKSPSSRestrictedLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSPSSRestrictedLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSPSSRestrictedLab) EstimatedTime() int     { return 25 }
func (l *CKSPSSRestrictedLab) Cert() Cert             { return CertCKS }
func (l *CKSPSSRestrictedLab) DomainWeight() int      { return 20 }
func (l *CKSPSSRestrictedLab) Tags() []string {
	return []string{"cks", "pod-security", "restricted", "microservice-vulns"}
}

func (l *CKSPSSRestrictedLab) Description() string {
	return `The 'restricted-ns' namespace has no Pod Security Standards enforcement.
Pods running with privileged containers or running as root are allowed.

Your task: Label the 'restricted-ns' namespace to enforce the 'restricted'
Pod Security Standard at the 'enforce' level. Then fix any pods that violate
the policy.`
}

func (l *CKSPSSRestrictedLab) Hints() []string {
	return []string{
		"Use kubectl label namespace to set Pod Security Standards",
		"Label: pod-security.kubernetes.io/enforce=restricted",
		"Delete and recreate violating pods with proper security context",
	}
}

func (l *CKSPSSRestrictedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSPSSRestrictedLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: restricted-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: insecure-pod
  namespace: restricted-ns
spec:
  containers:
  - name: app
    image: nginx:alpine
    securityContext:
      privileged: true
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSPSSRestrictedLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "namespace", "restricted-ns",
		"-o", "jsonpath={.metadata.labels}")
	if err != nil {
		return fmt.Errorf("failed to get namespace: %w", err)
	}
	if !strings.Contains(output, "restricted") {
		return fmt.Errorf("restricted PSS not applied")
	}
	return nil
}

func (l *CKSPSSRestrictedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Label namespace for restricted PSS", Command: "kubectl label namespace restricted-ns pod-security.kubernetes.io/enforce=restricted pod-security.kubernetes.io/enforce-version=latest"},
		{Description: "Delete violating pod", Command: "kubectl delete pod insecure-pod -n restricted-ns"},
		{Description: "Recreate with compliant spec", Command: `kubectl run secure-pod -n restricted-ns --image=nginx:alpine --restart=Never --overrides='{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"seccompProfile":{"type":"RuntimeDefault"}},"containers":[{"name":"secure-pod","image":"nginx:alpine","securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}}]}}'`},
		{Description: "Verify pod runs", Command: "kubectl get pod secure-pod -n restricted-ns"},
	}
}
