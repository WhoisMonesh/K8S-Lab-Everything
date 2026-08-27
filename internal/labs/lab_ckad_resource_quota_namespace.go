package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADResourceQuotaNamespaceLab{})
}

type CKADResourceQuotaNamespaceLab struct {
	BaseLab
}

func (l *CKADResourceQuotaNamespaceLab) ID() string {
	return "ckad_resource_quota_namespace"
}

func (l *CKADResourceQuotaNamespaceLab) Title() string {
	return "Create ResourceQuota for Namespace"
}

func (l *CKADResourceQuotaNamespaceLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADResourceQuotaNamespaceLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADResourceQuotaNamespaceLab) Cert() Cert             { return CertCKAD }
func (l *CKADResourceQuotaNamespaceLab) DomainWeight() int      { return 25 }
func (l *CKADResourceQuotaNamespaceLab) EstimatedTime() int     { return 15 }
func (l *CKADResourceQuotaNamespaceLab) Tags() []string {
	return []string{"resource-quota", "namespace", "limits"}
}

func (l *CKADResourceQuotaNamespaceLab) Description() string {
	return `A namespace needs resource limits to prevent resource exhaustion. Create
a ResourceQuota that limits CPU, memory, and pod count.

Your task: Create a ResourceQuota for the 'limited' namespace.`
}

func (l *CKADResourceQuotaNamespaceLab) Hints() []string {
	return []string{
		"Create the namespace first if it doesn't exist",
		"Use hard limits for CPU, memory, and pod count",
		"Specify requests.cpu, requests.memory, limits.cpu, limits.memory",
	}
}

func (l *CKADResourceQuotaNamespaceLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADResourceQuotaNamespaceLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: limited`
	return kubectlApply(ctx, kubeconfigPath, ns)
}

func (l *CKADResourceQuotaNamespaceLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "resourcequota", "-n", "limited",
		"-o", "jsonpath={.items[0].spec.hard}")
	if err != nil {
		return fmt.Errorf("failed to get resourcequota: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no ResourceQuota found")
	}
	return nil
}

func (l *CKADResourceQuotaNamespaceLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create namespace", Command: "kubectl create namespace limited"},
		{Description: "Create ResourceQuota", Command: "Create ResourceQuota with hard limits for cpu, memory, and pods"},
		{Description: "Verify", Command: "kubectl get resourcequota -n limited"},
	}
}
