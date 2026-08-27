package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&RuntimeClassNotFoundLab{})
}

type RuntimeClassNotFoundLab struct {
	BaseLab
}

func (l *RuntimeClassNotFoundLab) ID() string {
	return "runtime_class_not_found"
}

func (l *RuntimeClassNotFoundLab) Title() string {
	return "RuntimeClass Not Found"
}

func (l *RuntimeClassNotFoundLab) Category() Category {
	return CategoryScheduling
}

func (l *RuntimeClassNotFoundLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *RuntimeClassNotFoundLab) Description() string {
	return `A pod 'isolated-app' references a RuntimeClass 'kata' that doesn't
exist in the cluster. The pod cannot be scheduled because the RuntimeClass
is not found.

Your task: Either create the RuntimeClass or remove the reference.`
}

func (l *RuntimeClassNotFoundLab) Hints() []string {
	return []string{
		"Check the pod's runtimeClassName",
		"List available RuntimeClasses",
		"Either create the RuntimeClass or remove the field",
	}
}

func (l *RuntimeClassNotFoundLab) EstimatedTime() int {
	return 10
}

func (l *RuntimeClassNotFoundLab) Tags() []string {
	return []string{"runtimeclass", "scheduling"}
}

func (l *RuntimeClassNotFoundLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *RuntimeClassNotFoundLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: isolated-app
  namespace: default
spec:
  runtimeClassName: kata
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'sleep 3600']
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *RuntimeClassNotFoundLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pod", "isolated-app",
		"-o", "jsonpath={.status.phase}")
	if strings.TrimSpace(output) == "Pending" {
		return nil
	}
	return nil
}

func (l *RuntimeClassNotFoundLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "isolated-app",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *RuntimeClassNotFoundLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod RuntimeClass",
			Command:     "kubectl get pod isolated-app -o yaml | grep runtimeClassName",
			Notes:       "runtimeClassName is 'kata'",
		},
		{
			Description: "Check available RuntimeClasses",
			Command:     "kubectl get runtimeclasses",
			Notes:       "No RuntimeClass named 'kata' exists",
		},
		{
			Description: "Option A: Remove RuntimeClass",
			Command:     "kubectl edit pod isolated-app",
			Notes:       "Remove the runtimeClassName field",
		},
		{
			Description: "Option B: Create RuntimeClass",
			Command:     "kubectl apply -f - <<EOF\napiVersion: node.k8s.io/v1\nkind: RuntimeClass\nmetadata:\n  name: kata\nhandler: kata\nEOF",
			Notes:       "Create the RuntimeClass if kata runtime is available",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod isolated-app",
			Notes:       "Pod should now be Running",
		},
	}
}
