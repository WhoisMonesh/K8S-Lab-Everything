package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ResourceQuotaScopePodLab{})
}

type ResourceQuotaScopePodLab struct {
	BaseLab
}

func (l *ResourceQuotaScopePodLab) ID() string {
	return "resource_quota_scope_pod"
}

func (l *ResourceQuotaScopePodLab) Title() string {
	return "ResourceQuota Scope Excluding Pods"
}

func (l *ResourceQuotaScopePodLab) Category() Category {
	return CategoryScheduling
}

func (l *ResourceQuotaScopePodLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ResourceQuotaScopePodLab) Description() string {
	return `A ResourceQuota 'compute-quota' has scope: Restricted which excludes
pods from quota calculation. The deployment can't create pods because
the quota calculation doesn't account for pod resources correctly.

Your task: Fix the ResourceQuota scope to properly include pods.`
}

func (l *ResourceQuotaScopePodLab) Hints() []string {
	return []string{
		"Check the ResourceQuota scopes",
		"Restricted scope excludes pods from quota",
		"Remove the scope restriction or use NotTerminating scope",
	}
}

func (l *ResourceQuotaScopePodLab) EstimatedTime() int {
	return 15
}

func (l *ResourceQuotaScopePodLab) Tags() []string {
	return []string{"resourcequota", "scope", "scheduling"}
}

func (l *ResourceQuotaScopePodLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ResourceQuotaScopePodLab) Break(ctx context.Context, kubeconfigPath string) error {
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: quota-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	quota := `apiVersion: v1
kind: ResourceQuota
metadata:
  name: compute-quota
  namespace: quota-ns
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 8Gi
    limits.cpu: "8"
    limits.memory: 16Gi
  scopeSelector:
    matchExpressions:
    - scopeName: Restricted
      operator: In
`
	if err := kubectlApply(ctx, kubeconfigPath, quota); err != nil {
		return fmt.Errorf("creating quota: %w", err)
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: quota-ns
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web-app
  template:
    metadata:
      labels:
        app: web-app
    spec:
      containers:
      - name: web
        image: nginx:alpine
        resources:
          requests:
            cpu: 500m
            memory: 256Mi
          limits:
            cpu: 1
            memory: 512Mi
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	return nil
}

func (l *ResourceQuotaScopePodLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *ResourceQuotaScopePodLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "web-app",
		"-n", "quota-ns", "-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	if strings.TrimSpace(output) != "3" {
		return fmt.Errorf("deployment not ready (ready replicas: %s, expected: 3)", output)
	}

	return nil
}

func (l *ResourceQuotaScopePodLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check ResourceQuota",
			Command:     "kubectl get resourcequota compute-quota -n quota-ns -o yaml",
			Notes:       "scopeSelector uses Restricted scope which excludes pods",
		},
		{
			Description: "Fix ResourceQuota scope",
			Command:     "kubectl edit resourcequota compute-quota -n quota-ns",
			Notes:       "Remove scopeSelector or change to NotTerminating scope",
		},
		{
			Description: "Verify deployment is ready",
			Command:     "kubectl get deployment web-app -n quota-ns",
			Notes:       "All 3 replicas should be ready",
		},
	}
}
