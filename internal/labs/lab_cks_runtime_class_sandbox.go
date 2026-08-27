package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSRuntimeClassSandboxLab{})
}

type CKSRuntimeClassSandboxLab struct {
	BaseLab
}

func (l *CKSRuntimeClassSandboxLab) ID() string             { return "cks_runtime_class_sandbox" }
func (l *CKSRuntimeClassSandboxLab) Title() string          { return "Use RuntimeClass with Sandbox" }
func (l *CKSRuntimeClassSandboxLab) Category() Category     { return CategoryMicroserviceVulns }
func (l *CKSRuntimeClassSandboxLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSRuntimeClassSandboxLab) EstimatedTime() int     { return 20 }
func (l *CKSRuntimeClassSandboxLab) Cert() Cert             { return CertCKS }
func (l *CKSRuntimeClassSandboxLab) DomainWeight() int      { return 20 }
func (l *CKSRuntimeClassSandboxLab) Tags() []string {
	return []string{"cks", "runtime-class", "sandbox", "gvisor", "microservice-vulns"}
}

func (l *CKSRuntimeClassSandboxLab) Description() string {
	return `The 'sandboxed' namespace has pods running with the default runtime.
For enhanced security, pods should use a sandboxed runtime like gVisor.

Your task: Create a RuntimeClass 'gvisor' configured for gVisor and then
create a pod that uses this RuntimeClass.`
}

func (l *CKSRuntimeClassSandboxLab) Hints() []string {
	return []string{
		"Create a RuntimeClass with handler: gvisor",
		"Set the scheduling nodeSelector to the runtime class",
		"Reference the RuntimeClass in the pod spec",
	}
}

func (l *CKSRuntimeClassSandboxLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSRuntimeClassSandboxLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: sandboxed
`
	return kubectlApply(ctx, kubeconfigPath, ns)
}

func (l *CKSRuntimeClassSandboxLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "runtimeclass", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get runtime classes: %w", err)
	}
	if strings.Contains(output, "gvisor") {
		return nil
	}
	return fmt.Errorf("gvisor RuntimeClass not found")
}

func (l *CKSRuntimeClassSandboxLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create RuntimeClass", Command: `cat <<EOF | kubectl apply -f -
apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: gvisor
handler: gvisor
scheduling:
  nodeSelector:
    node.kubernetes.io/runtime-class: gvisor
EOF`},
		{Description: "Create pod using RuntimeClass", Command: `kubectl run sandboxed-pod -n sandboxed --image=busybox:1.36 --restart=Never --overrides='{"spec":{"runtimeClassName":"gvisor","containers":[{"name":"sandboxed-pod","image":"busybox:1.36","command":["sh","-c","while true; do sleep 3600; done"]}]}}'`},
		{Description: "Verify", Command: "kubectl get runtimeclass gvisor"},
	}
}
