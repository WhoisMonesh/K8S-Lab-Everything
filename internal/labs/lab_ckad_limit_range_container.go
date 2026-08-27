package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADLimitRangeContainerLab{})
}

type CKADLimitRangeContainerLab struct {
	BaseLab
}

func (l *CKADLimitRangeContainerLab) ID() string {
	return "ckad_limit_range_container"
}

func (l *CKADLimitRangeContainerLab) Title() string {
	return "Create LimitRange for Containers"
}

func (l *CKADLimitRangeContainerLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADLimitRangeContainerLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADLimitRangeContainerLab) Cert() Cert             { return CertCKAD }
func (l *CKADLimitRangeContainerLab) DomainWeight() int      { return 25 }
func (l *CKADLimitRangeContainerLab) EstimatedTime() int     { return 15 }
func (l *CKADLimitRangeContainerLab) Tags() []string {
	return []string{"limit-range", "defaults", "constraints"}
}

func (l *CKADLimitRangeContainerLab) Description() string {
	return `A namespace needs default resource limits for containers. Create a LimitRange
that sets default CPU and memory limits for all containers.

Your task: Create a LimitRange in the 'default-limits' namespace.`
}

func (l *CKADLimitRangeContainerLab) Hints() []string {
	return []string{
		"Use type Container in the LimitRange spec",
		"Set default and defaultRequest limits",
		"Containers without explicit limits will inherit defaults",
	}
}

func (l *CKADLimitRangeContainerLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADLimitRangeContainerLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: default-limits`
	return kubectlApply(ctx, kubeconfigPath, ns)
}

func (l *CKADLimitRangeContainerLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "limitrange", "-n", "default-limits",
		"-o", "jsonpath={.items[0].spec.limits[*].type}")
	if err != nil {
		return fmt.Errorf("failed to get limitrange: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no LimitRange found")
	}
	if !strings.Contains(output, "Container") {
		return fmt.Errorf("LimitRange doesn't have Container type")
	}
	return nil
}

func (l *CKADLimitRangeContainerLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create namespace", Command: "kubectl create namespace default-limits"},
		{Description: "Create LimitRange", Command: "Create LimitRange with Container type and default limits"},
		{Description: "Verify", Command: "kubectl get limitrange -n default-limits"},
	}
}
