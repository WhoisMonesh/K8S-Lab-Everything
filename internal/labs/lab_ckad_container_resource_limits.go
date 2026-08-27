package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADContainerResourceLimitsLab{})
}

type CKADContainerResourceLimitsLab struct {
	BaseLab
}

func (l *CKADContainerResourceLimitsLab) ID() string             { return "ckad_container_resource_limits" }
func (l *CKADContainerResourceLimitsLab) Title() string          { return "Set Resource Limits" }
func (l *CKADContainerResourceLimitsLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADContainerResourceLimitsLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADContainerResourceLimitsLab) Cert() Cert             { return CertCKAD }
func (l *CKADContainerResourceLimitsLab) DomainWeight() int      { return 20 }
func (l *CKADContainerResourceLimitsLab) EstimatedTime() int     { return 10 }
func (l *CKADContainerResourceLimitsLab) Tags() []string {
	return []string{"resource-limits", "resources", "quotas"}
}

func (l *CKADContainerResourceLimitsLab) Description() string {
	return `A pod is consuming too many resources, affecting other workloads on the node.

Your task: Set appropriate resource limits for the pod containers to prevent
resource exhaustion.`
}

func (l *CKADContainerResourceLimitsLab) Hints() []string {
	return []string{
		"Resource limits prevent containers from using more than specified resources",
		"Set cpu and memory limits in the container spec",
		"Exceeding memory limits causes OOMKill",
	}
}

func (l *CKADContainerResourceLimitsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADContainerResourceLimitsLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: resource-hog
  labels:
    app: resource-hog
spec:
  containers:
  - name: app
    image: nginx:alpine
    resources: {}`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADContainerResourceLimitsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "resource-hog",
		"-o", "jsonpath={.spec.containers[0].resources.limits}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no resource limits set")
	}
	return nil
}

func (l *CKADContainerResourceLimitsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod to add resource limits", Command: "kubectl edit pod resource-hog"},
		{Description: "Add limits section", Command: "Add resources.limits with cpu: 500m and memory: 256Mi"},
		{Description: "Verify limits", Command: "kubectl get pod resource-hog -o yaml | grep -A 3 limits"},
	}
}
