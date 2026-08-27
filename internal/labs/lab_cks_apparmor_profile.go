package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSAppArmorProfileLab{})
}

type CKSAppArmorProfileLab struct {
	BaseLab
}

func (l *CKSAppArmorProfileLab) ID() string             { return "cks_apparmor_profile" }
func (l *CKSAppArmorProfileLab) Title() string          { return "Configure AppArmor Profile for Pods" }
func (l *CKSAppArmorProfileLab) Category() Category     { return CategorySystemHardening }
func (l *CKSAppArmorProfileLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSAppArmorProfileLab) EstimatedTime() int     { return 20 }
func (l *CKSAppArmorProfileLab) Cert() Cert             { return CertCKS }
func (l *CKSAppArmorProfileLab) DomainWeight() int      { return 10 }
func (l *CKSAppArmorProfileLab) Tags() []string {
	return []string{"cks", "apparmor", "pod-security", "system-hardening"}
}

func (l *CKSAppArmorProfileLab) Description() string {
	return `A pod 'secure-pod' in namespace 'hardened' runs without an AppArmor profile.
This allows unrestricted access to system resources.

Your task: Annotate the pod to use the 'docker-default' AppArmor profile
and ensure the pod runs successfully.`
}

func (l *CKSAppArmorProfileLab) Hints() []string {
	return []string{
		"Use pod annotations for AppArmor",
		"Set container.apparmor.security.beta.kubernetes.io/<container>: <profile>",
		"Delete and recreate the pod with the annotation",
	}
}

func (l *CKSAppArmorProfileLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSAppArmorProfileLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: hardened
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: secure-pod
  namespace: hardened
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSAppArmorProfileLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "secure-pod", "-n", "hardened",
		"-o", "jsonpath={.metadata.annotations}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.Contains(output, "apparmor") {
		return nil
	}
	return fmt.Errorf("AppArmor annotation not set")
}

func (l *CKSAppArmorProfileLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Delete and recreate pod with AppArmor", Command: `kubectl delete pod secure-pod -n hardened && kubectl run secure-pod -n hardened --image=busybox:1.36 --restart=Never --overrides='{"metadata":{"annotations":{"container.apparmor.security.beta.kubernetes.io/app":"docker-default"}},"spec":{"containers":[{"name":"secure-pod","image":"busybox:1.36","command":["sh","-c","while true; do sleep 3600; done"]}]}}'`},
		{Description: "Verify AppArmor annotation", Command: "kubectl get pod secure-pod -n hardened -o jsonpath='{.metadata.annotations}'"},
	}
}
