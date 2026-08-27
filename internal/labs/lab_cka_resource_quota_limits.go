package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&ResourceQuotaLimitsLab{})
}

type ResourceQuotaLimitsLab struct {
	BaseLab
}

func (l *ResourceQuotaLimitsLab) ID() string { return "cka_resource_quota_limits" }
func (l *ResourceQuotaLimitsLab) Title() string {
	return "Set Resource Quotas and Limits"
}
func (l *ResourceQuotaLimitsLab) Category() Category     { return CategoryWorkloadsScheduling }
func (l *ResourceQuotaLimitsLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *ResourceQuotaLimitsLab) EstimatedTime() int     { return 20 }
func (l *ResourceQuotaLimitsLab) Tags() []string {
	return []string{"resourcequota", "limits", "scheduling"}
}
func (l *ResourceQuotaLimitsLab) Cert() Cert        { return CertCKA }
func (l *ResourceQuotaLimitsLab) DomainWeight() int { return 15 }

func (l *ResourceQuotaLimitsLab) Description() string {
	return `A namespace has no resource quotas, allowing pods to consume excessive
resources. Create a ResourceQuota limiting CPU, memory, and pod count, and
a LimitRange to set default limits for containers.`
}

func (l *ResourceQuotaLimitsLab) Hints() []string {
	return []string{
		"Create a ResourceQuota in the namespace",
		"Set hard limits for cpu, memory, and pods",
		"Create a LimitRange for default container limits",
	}
}

func (l *ResourceQuotaLimitsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ResourceQuotaLimitsLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *ResourceQuotaLimitsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "resourcequota", "-n", "quota-ns",
		"-o", "name")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "resourcequota") {
		return fmt.Errorf("ResourceQuota not created")
	}
	return nil
}

func (l *ResourceQuotaLimitsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create ResourceQuota", Command: "cat <<EOF | kubectl apply -f -\napiVersion: v1\nkind: ResourceQuota\nmetadata:\n  name: compute-quota\n  namespace: quota-ns\nspec:\n  hard:\n    requests.cpu: \"4\"\n    requests.memory: 8Gi\n    limits.cpu: \"8\"\n    limits.memory: 16Gi\n    pods: \"10\"\nEOF"},
		{Description: "Create LimitRange", Command: "cat <<EOF | kubectl apply -f -\napiVersion: v1\nkind: LimitRange\nmetadata:\n  name: default-limits\n  namespace: quota-ns\nspec:\n  limits:\n  - default:\n      cpu: 500m\n      memory: 256Mi\n    defaultRequest:\n      cpu: 100m\n      memory: 128Mi\n    type: Container\nEOF"},
		{Description: "Verify", Command: "kubectl get resourcequota -n quota-ns"},
	}
}
