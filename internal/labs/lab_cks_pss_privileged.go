package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSPSSPrivilegedLab{})
}

type CKSPSSPrivilegedLab struct {
	BaseLab
}

func (l *CKSPSSPrivilegedLab) ID() string             { return "cks_pss_privileged" }
func (l *CKSPSSPrivilegedLab) Title() string          { return "Block Privileged Containers" }
func (l *CKSPSSPrivilegedLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSPSSPrivilegedLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSPSSPrivilegedLab) EstimatedTime() int     { return 20 }
func (l *CKSPSSPrivilegedLab) Cert() Cert             { return CertCKS }
func (l *CKSPSSPrivilegedLab) DomainWeight() int      { return 20 }
func (l *CKSPSSPrivilegedLab) Tags() []string {
	return []string{"cks", "privileged", "pod-security", "microservice-vulns"}
}

func (l *CKSPSSPrivilegedLab) Description() string {
	return `A pod 'privil-worker' in namespace 'no-priv' runs with privileged: true,
which gives it full access to the host.

Your task: Remove the privileged flag from the pod and ensure it runs with
minimal capabilities.`
}

func (l *CKSPSSPrivilegedLab) Hints() []string {
	return []string{
		"Check the pod spec for securityContext.privileged",
		"Set privileged to false",
		"Add capabilities drop ALL for additional hardening",
	}
}

func (l *CKSPSSPrivilegedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSPSSPrivilegedLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: no-priv
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: privil-worker
  namespace: no-priv
spec:
  containers:
  - name: worker
    image: busybox:1.36
    securityContext:
      privileged: true
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSPSSPrivilegedLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "privil-worker", "-n", "no-priv",
		"-o", "jsonpath={.spec.containers[0].securityContext.privileged}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "true" {
		return fmt.Errorf("pod still privileged")
	}
	return nil
}

func (l *CKSPSSPrivilegedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Delete privileged pod", Command: "kubectl delete pod privil-worker -n no-priv"},
		{Description: "Recreate with non-privileged security context", Command: `kubectl run privil-worker -n no-priv --image=busybox:1.36 --restart=Never --overrides='{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"seccompProfile":{"type":"RuntimeDefault"}},"containers":[{"name":"privil-worker","image":"busybox:1.36","securityContext":{"privileged":false,"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}},"command":["sh","-c","while true; do sleep 3600; done"]}]}}'`},
		{Description: "Verify pod runs", Command: "kubectl get pod privil-worker -n no-priv"},
	}
}
