package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&CKSKernelHardeningLab{})
}

type CKSKernelHardeningLab struct {
	BaseLab
}

func (l *CKSKernelHardeningLab) ID() string             { return "cks_kernel_hardening" }
func (l *CKSKernelHardeningLab) Title() string          { return "Configure Kernel Hardening Parameters" }
func (l *CKSKernelHardeningLab) Category() Category     { return CategorySystemHardening }
func (l *CKSKernelHardeningLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSKernelHardeningLab) EstimatedTime() int     { return 25 }
func (l *CKSKernelHardeningLab) Cert() Cert             { return CertCKS }
func (l *CKSKernelHardeningLab) DomainWeight() int      { return 10 }
func (l *CKSKernelHardeningLab) Tags() []string {
	return []string{"cks", "kernel", "sysctl", "hardening"}
}

func (l *CKSKernelHardeningLab) Description() string {
	return `A pod 'worker' in namespace 'kernel-hardening' is running with hostPID: true
and a privileged securityContext. This exposes host namespaces and allows the
container to access host/node resources.

Your task: Harden the pod so that it no longer shares the host PID namespace
(hostPID removed/disabled) and is not privileged. For extra hardening, also
make the root filesystem read-only.`
}

func (l *CKSKernelHardeningLab) Hints() []string {
	return []string{
		"Remove hostPID: true from the pod spec",
		"Set securityContext.privileged to false",
		"Set readOnlyRootFilesystem to true for extra hardening",
	}
}

func (l *CKSKernelHardeningLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSKernelHardeningLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: kernel-hardening
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: worker
  namespace: kernel-hardening
spec:
  hostPID: true
  containers:
  - name: worker
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
    securityContext:
      privileged: true
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSKernelHardeningLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *CKSKernelHardeningLab) Verify(ctx context.Context, kubeconfigPath string) error {
	hostPID, err := kubectl(ctx, kubeconfigPath, "get", "pod", "worker", "-n", "kernel-hardening",
		"-o", "jsonpath={.spec.hostPID}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(hostPID) == "true" {
		return fmt.Errorf("hostPID is still enabled on the pod")
	}
	priv, err := kubectl(ctx, kubeconfigPath, "get", "pod", "worker", "-n", "kernel-hardening",
		"-o", "jsonpath={.spec.securityContext.privileged}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(priv) == "true" {
		return fmt.Errorf("pod is still privileged")
	}
	return nil
}

func (l *CKSKernelHardeningLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Delete and recreate pod without hostPID/privileged", Command: `kubectl delete pod worker -n kernel-hardening && kubectl run worker -n kernel-hardening --image=busybox:1.36 --restart=Never --overrides='{"spec":{"containers":[{"name":"worker","image":"busybox:1.36","command":["sh","-c","while true; do sleep 3600; done"],"securityContext":{"privileged":false,"readOnlyRootFilesystem":true}}]}}'`},
		{Description: "Verify hostPID and privileged are gone", Command: "kubectl get pod worker -n kernel-hardening -o jsonpath='{.spec.hostPID} {.spec.securityContext.privileged}'"},
	}
}
