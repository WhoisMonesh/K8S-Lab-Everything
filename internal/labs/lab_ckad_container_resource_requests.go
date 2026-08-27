package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADContainerResourceRequestsLab{})
}

type CKADContainerResourceRequestsLab struct {
	BaseLab
}

func (l *CKADContainerResourceRequestsLab) ID() string             { return "ckad_container_resource_requests" }
func (l *CKADContainerResourceRequestsLab) Title() string          { return "Set Resource Requests" }
func (l *CKADContainerResourceRequestsLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADContainerResourceRequestsLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADContainerResourceRequestsLab) Cert() Cert             { return CertCKAD }
func (l *CKADContainerResourceRequestsLab) DomainWeight() int      { return 20 }
func (l *CKADContainerResourceRequestsLab) EstimatedTime() int     { return 10 }
func (l *CKADContainerResourceRequestsLab) Tags() []string {
	return []string{"resource-requests", "scheduling", "resources"}
}

func (l *CKADContainerResourceRequestsLab) Description() string {
	return `A pod is being scheduled on nodes with insufficient resources,
causing performance issues.

Your task: Set appropriate resource requests for the pod containers.`
}

func (l *CKADContainerResourceRequestsLab) Hints() []string {
	return []string{
		"Resource requests are used for scheduling decisions",
		"Set cpu and memory requests in the container spec",
		"Use units like 100m for CPU and 128Mi for memory",
	}
}

func (l *CKADContainerResourceRequestsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADContainerResourceRequestsLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: resource-app
  labels:
    app: resource-app
spec:
  containers:
  - name: app
    image: nginx:alpine
    resources: {}`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADContainerResourceRequestsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "resource-app",
		"-o", "jsonpath={.spec.containers[0].resources.requests}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no resource requests set")
	}
	return nil
}

func (l *CKADContainerResourceRequestsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod to add resource requests", Command: "kubectl edit pod resource-app"},
		{Description: "Add requests section", Command: "Add resources.requests with cpu: 100m and memory: 128Mi"},
		{Description: "Verify requests", Command: "kubectl get pod resource-app -o yaml | grep -A 3 requests"},
	}
}
