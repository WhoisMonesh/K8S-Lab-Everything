package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ResourceQuotaScopeLab{})
}

type ResourceQuotaScopeLab struct {
	BaseLab
}

func (l *ResourceQuotaScopeLab) ID() string {
	return "resource_quota_scope"
}

func (l *ResourceQuotaScopeLab) Title() string {
	return "Resource Quota Scope Limitation"
}

func (l *ResourceQuotaScopeLab) Category() Category {
	return CategoryAppConfigSecurity
}

func (l *ResourceQuotaScopeLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ResourceQuotaScopeLab) Description() string {
	return `A ResourceQuota is configured but is applied to all pods including
Burstable/QoS classes. Only BestEffort pods should be limited by
this quota.

Your task: Configure the ResourceQuota with a scope selector to only
apply to BestEffort pods.`
}

func (l *ResourceQuotaScopeLab) Hints() []string {
	return []string{
		"Check the current ResourceQuota configuration",
		"Use spec.scopeSelector with matchingScopes",
		"BestEffort scope limits pods with no resource requests/limits",
	}
}

func (l *ResourceQuotaScopeLab) EstimatedTime() int {
	return 20
}

func (l *ResourceQuotaScopeLab) Tags() []string {
	return []string{"resource-quota", "scope", "besteffort", "qos", "security"}
}

func (l *ResourceQuotaScopeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ResourceQuotaScopeLab) Break(ctx context.Context, kubeconfigPath string) error {
	quota := `apiVersion: v1
kind: ResourceQuota
metadata:
  name: besteffort-quota
  namespace: default
spec:
  hard:
    pods: "3"
`
	return kubectlApply(ctx, kubeconfigPath, quota)
}

func (l *ResourceQuotaScopeLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(5 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "resourcequota", "besteffort-quota",
		"-o", "jsonpath={.spec.scopeSelector}")
	if err != nil {
		return nil
	}

	if strings.TrimSpace(output) == "" || strings.TrimSpace(output) == "<none>" {
		return nil
	}

	return fmt.Errorf("quota has scope selector (expected none)")
}

func (l *ResourceQuotaScopeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(5 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "resourcequota", "besteffort-quota",
		"-o", "jsonpath={.spec.scopeSelector}")
	if err != nil {
		return fmt.Errorf("checking quota: %w", err)
	}

	if strings.TrimSpace(output) == "" || strings.TrimSpace(output) == "<none>" {
		return fmt.Errorf("ResourceQuota still has no scope selector")
	}

	return nil
}

func (l *ResourceQuotaScopeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check current quota",
			Command:     "kubectl get resourcequota besteffort-quota -o yaml",
			Notes:       "No scopeSelector - applies to all pods",
		},
		{
			Description: "Fix: Add scope selector for BestEffort",
			Command:     `kubectl patch resourcequota besteffort-quota --type='merge' -p '{"spec":{"scopeSelector":{"matchExpressions":[{"scopeName":"BestEffort","operator":"In"}]}}}'`,
			Notes:       "Only limit BestEffort pods",
		},
		{
			Description: "Verify scope selector is set",
			Command:     "kubectl get resourcequota besteffort-quota -o yaml | grep -A 5 scopeSelector",
			Notes:       "Should show BestEffort scope",
		},
	}
}
