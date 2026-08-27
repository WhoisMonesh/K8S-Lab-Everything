package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSPSSBaselineLab{})
}

type CKSPSSBaselineLab struct {
	BaseLab
}

func (l *CKSPSSBaselineLab) ID() string             { return "cks_pss_baseline" }
func (l *CKSPSSBaselineLab) Title() string          { return "Enforce Pod Security Standards Baseline" }
func (l *CKSPSSBaselineLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSPSSBaselineLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSPSSBaselineLab) EstimatedTime() int     { return 20 }
func (l *CKSPSSBaselineLab) Cert() Cert             { return CertCKS }
func (l *CKSPSSBaselineLab) DomainWeight() int      { return 20 }
func (l *CKSPSSBaselineLab) Tags() []string {
	return []string{"cks", "pod-security", "baseline", "microservice-vulns"}
}

func (l *CKSPSSBaselineLab) Description() string {
	return `The 'baseline-ns' namespace does not enforce any Pod Security Standards.
A pod with hostNetwork=true is running, which violates the baseline profile.

Your task: Label the namespace to enforce the 'baseline' Pod Security Standard
and ensure the violating pod is corrected.`
}

func (l *CKSPSSBaselineLab) Hints() []string {
	return []string{
		"Label namespace: pod-security.kubernetes.io/enforce=baseline",
		"Remove hostNetwork: true from the pod spec",
		"Delete and recreate the pod",
	}
}

func (l *CKSPSSBaselineLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSPSSBaselineLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: baseline-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: host-net-pod
  namespace: baseline-ns
spec:
  hostNetwork: true
  containers:
  - name: app
    image: nginx:alpine
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSPSSBaselineLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "namespace", "baseline-ns",
		"-o", "jsonpath={.metadata.labels}")
	if err != nil {
		return fmt.Errorf("failed to get namespace: %w", err)
	}
	if !strings.Contains(output, "baseline") {
		return fmt.Errorf("baseline PSS not applied")
	}

	podOutput, err := kubectl(ctx, kubeconfigPath, "get", "pod", "host-net-pod", "-n", "baseline-ns",
		"-o", "jsonpath={.spec.hostNetwork}")
	if err == nil && strings.TrimSpace(podOutput) == "true" {
		return fmt.Errorf("pod still has hostNetwork=true")
	}
	return nil
}

func (l *CKSPSSBaselineLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Label namespace for baseline PSS", Command: "kubectl label namespace baseline-ns pod-security.kubernetes.io/enforce=baseline"},
		{Description: "Delete violating pod", Command: "kubectl delete pod host-net-pod -n baseline-ns"},
		{Description: "Recreate without hostNetwork", Command: "kubectl run host-net-pod -n baseline-ns --image=nginx:alpine --restart=Never"},
		{Description: "Verify pod runs", Command: "kubectl get pod host-net-pod -n baseline-ns"},
	}
}
