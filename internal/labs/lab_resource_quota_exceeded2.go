package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ResourceQuotaExceeded{})
}

type ResourceQuotaExceeded struct {
	BaseLab
}

func (l *ResourceQuotaExceeded) ID() string             { return "resource_quota_exceeded2" }
func (l *ResourceQuotaExceeded) Title() string          { return "Pods Rejected by ResourceQuota" }
func (l *ResourceQuotaExceeded) Category() Category     { return CategoryWorkloads }
func (l *ResourceQuotaExceeded) Difficulty() Difficulty { return DifficultyMedium }
func (l *ResourceQuotaExceeded) EstimatedTime() int     { return 20 }
func (l *ResourceQuotaExceeded) Tags() []string {
	return []string{"resourcequota", "limits", "scheduling"}
}

func (l *ResourceQuotaExceeded) Description() string {
	return `Pods are being rejected because the namespace ResourceQuota has been exceeded.
Clean up resources or increase the quota limits.`
}

func (l *ResourceQuotaExceeded) Hints() []string {
	return []string{
		"Check ResourceQuota in the namespace",
		"Look at resource usage vs limits",
		"Increase the quota or reduce usage",
	}
}

func (l *ResourceQuotaExceeded) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ResourceQuotaExceeded) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: quota-test
---
apiVersion: v1
kind: ResourceQuota
metadata:
  name: compute-quota
  namespace: quota-test
spec:
  hard:
    requests.cpu: "1"
    requests.memory: "512Mi"
    limits.cpu: "2"
    limits.memory: "1Gi"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: big-app
  namespace: quota-test
spec:
  replicas: 5
  selector:
    matchLabels:
      app: big-app
  template:
    metadata:
      labels:
        app: big-app
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        resources:
          requests:
            cpu: "500m"
            memory: "256Mi"
          limits:
            cpu: "1"
            memory: "512Mi"`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ResourceQuotaExceeded) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "resourcequota", "compute-quota", "-n", "quota-test",
		"-o", "jsonpath={.status.used}")
	if err != nil {
		return err
	}
	if output == "" {
		return fmt.Errorf("quota not configured")
	}
	return nil
}

func (l *ResourceQuotaExceeded) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check quota", Command: "kubectl get resourcequota -n quota-test"},
		{Description: "Check usage", Command: "kubectl describe resourcequota compute-quota -n quota-test"},
		{Description: "Increase quota", Command: "kubectl patch resourcequota compute-quota -n quota-test -p '{\"spec\":{\"hard\":{\"requests.cpu\":\"4\",\"limits.cpu\":\"8\",\"requests.memory\":\"2Gi\",\"limits.memory\":\"4Gi\"}}}'"},
	}
}
