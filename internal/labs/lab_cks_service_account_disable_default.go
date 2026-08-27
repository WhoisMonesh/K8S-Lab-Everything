package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSServiceAccountDisableDefaultLab{})
}

type CKSServiceAccountDisableDefaultLab struct {
	BaseLab
}

func (l *CKSServiceAccountDisableDefaultLab) ID() string {
	return "cks_service_account_disable_default"
}
func (l *CKSServiceAccountDisableDefaultLab) Title() string          { return "Disable Default ServiceAccount" }
func (l *CKSServiceAccountDisableDefaultLab) Category() Category     { return CategoryClusterHardening }
func (l *CKSServiceAccountDisableDefaultLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKSServiceAccountDisableDefaultLab) EstimatedTime() int     { return 15 }
func (l *CKSServiceAccountDisableDefaultLab) Cert() Cert             { return CertCKS }
func (l *CKSServiceAccountDisableDefaultLab) DomainWeight() int      { return 15 }
func (l *CKSServiceAccountDisableDefaultLab) Tags() []string {
	return []string{"cks", "service-account", "automount", "security"}
}

func (l *CKSServiceAccountDisableDefaultLab) Description() string {
	return `The default ServiceAccount in the 'restricted-ns' namespace has
automountServiceAccountToken set to true. This means all pods using the
default ServiceAccount can access the Kubernetes API.

Your task: Disable automountServiceAccountToken on the default ServiceAccount
in the 'restricted-ns' namespace and create a pod that explicitly disables it.`
}

func (l *CKSServiceAccountDisableDefaultLab) Hints() []string {
	return []string{
		"Use kubectl patch to update the ServiceAccount",
		"Set automountServiceAccountToken to false",
		"Also set it on individual pod specs",
	}
}

func (l *CKSServiceAccountDisableDefaultLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSServiceAccountDisableDefaultLab) Break(ctx context.Context, kubeconfigPath string) error {
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
  name: test-pod
  namespace: restricted-ns
spec:
  containers:
  - name: test
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSServiceAccountDisableDefaultLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "serviceaccount", "default", "-n", "restricted-ns",
		"-o", "jsonpath={.automountServiceAccountToken}")
	if err != nil {
		return fmt.Errorf("failed to check service account: %w", err)
	}
	if strings.TrimSpace(output) != "false" {
		return fmt.Errorf("automountServiceAccountToken not set to false on default SA")
	}

	podOutput, err := kubectl(ctx, kubeconfigPath, "get", "pod", "test-pod", "-n", "restricted-ns",
		"-o", "jsonpath={.spec.automountServiceAccountToken}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}
	if strings.TrimSpace(podOutput) != "false" {
		return fmt.Errorf("automountServiceAccountToken not set to false on pod")
	}
	return nil
}

func (l *CKSServiceAccountDisableDefaultLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Patch default ServiceAccount", Command: "kubectl patch serviceaccount default -n restricted-ns -p '{\"automountServiceAccountToken\": false}'"},
		{Description: "Update pod spec", Command: "kubectl delete pod test-pod -n restricted-ns && kubectl run test-pod -n restricted-ns --image=busybox:1.36 --restart=Never --overrides='{\"spec\":{\"automountServiceAccountToken\":false,\"containers\":[{\"name\":\"test-pod\",\"image\":\"busybox:1.36\",\"command\":[\"sh\",\"-c\",\"while true; do sleep 3600; done\"]}]}}'"},
		{Description: "Verify", Command: "kubectl get sa default -n restricted-ns -o jsonpath='{.automountServiceAccountToken}'"},
	}
}
