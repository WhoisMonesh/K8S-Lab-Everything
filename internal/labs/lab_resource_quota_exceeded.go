package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ResourceQuotaExceededLab{})
}

type ResourceQuotaExceededLab struct {
	BaseLab
}

func (l *ResourceQuotaExceededLab) ID() string {
	return "resource_quota_exceeded"
}

func (l *ResourceQuotaExceededLab) Title() string {
	return "Pods Rejected by ResourceQuota"
}

func (l *ResourceQuotaExceededLab) Category() Category {
	return CategoryWorkloads
}

func (l *ResourceQuotaExceededLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ResourceQuotaExceededLab) Description() string {
	return `A namespace 'team-alpha' has a ResourceQuota that's been exceeded.
New pods cannot be created because the quota is full.

Your task: Either adjust the quota or clean up resources to make room for new pods.`
}

func (l *ResourceQuotaExceededLab) Hints() []string {
	return []string{
		"Check the ResourceQuota status",
		"Look at how much quota is used vs available",
		"Check what resources are consuming the quota",
		"Either increase the quota or delete unused resources",
	}
}

func (l *ResourceQuotaExceededLab) EstimatedTime() int {
	return 20
}

func (l *ResourceQuotaExceededLab) Tags() []string {
	return []string{"resourcequota", "namespace", "resources", "scheduling"}
}

func (l *ResourceQuotaExceededLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ResourceQuotaExceededLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create namespace
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: team-alpha
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	// Create a tight ResourceQuota
	quota := `apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-quota
  namespace: team-alpha
spec:
  hard:
    requests.cpu: "2"
    requests.memory: "2Gi"
    limits.cpu: "4"
    limits.memory: "4Gi"
    pods: "5"
`
	if err := kubectlApply(ctx, kubeconfigPath, quota); err != nil {
		return fmt.Errorf("creating ResourceQuota: %w", err)
	}

	// Create pods that consume most of the quota
	for i := 0; i < 5; i++ {
		pod := fmt.Sprintf(`apiVersion: v1
kind: Pod
metadata:
  name: worker-%d
  namespace: team-alpha
  labels:
    app: worker
spec:
  containers:
  - name: app
    image: nginx:alpine
    resources:
      requests:
        cpu: "400m"
        memory: "400Mi"
      limits:
        cpu: "800m"
        memory: "800Mi"
`, i)
		if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
			return fmt.Errorf("creating pod: %w", err)
		}
	}

	return nil
}

func (l *ResourceQuotaExceededLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *ResourceQuotaExceededLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if a new pod can be created
	testPod := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: team-alpha
spec:
  containers:
  - name: test
    image: busybox:1.28
    command: ['sleep', '3600']
    resources:
      requests:
        cpu: "100m"
        memory: "100Mi"
`
	if err := kubectlApply(ctx, kubeconfigPath, testPod); err != nil {
		return fmt.Errorf("failed to create test pod: %w", err)
	}

	// Check if test pod was created
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "test-pod", "-n", "team-alpha",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check test pod: %w", err)
	}

	// Clean up test pod
	_, _ = kubectl(ctx, kubeconfigPath, "delete", "pod", "test-pod", "-n", "team-alpha")

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("test pod was not created (quota still exceeded)")
	}

	return nil
}

func (l *ResourceQuotaExceededLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check ResourceQuota status",
			Command:     "kubectl get resourcequota -n team-alpha",
			Notes:       "Shows used vs hard limits",
		},
		{
			Description: "Describe the ResourceQuota",
			Command:     "kubectl describe resourcequota team-quota -n team-alpha",
			Notes:       "See exactly how much of each resource is used",
		},
		{
			Description: "Check pods in the namespace",
			Command:     "kubectl get pods -n team-alpha",
			Notes:       "See all pods consuming resources",
		},
		{
			Description: "Option A: Delete unnecessary pods",
			Command:     "kubectl delete pod worker-0 worker-1 worker-2 -n team-alpha",
			Notes:       "Free up resources by removing pods",
		},
		{
			Description: "Option B: Increase the quota",
			Command:     "kubectl edit resourcequota team-quota -n team-alpha",
			Notes:       "Increase the CPU/memory limits in the quota",
		},
		{
			Description: "Verify new pods can be created",
			Command:     "kubectl run test-pod --image=busybox:1.28 -- sleep 3600 -n team-alpha",
			Notes:       "The pod should now be created successfully",
		},
	}
}
