package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADHelmCustomValuesLab{})
}

type CKADHelmCustomValuesLab struct {
	BaseLab
}

func (l *CKADHelmCustomValuesLab) ID() string             { return "ckad_helm_custom_values" }
func (l *CKADHelmCustomValuesLab) Title() string          { return "Customize Helm Values" }
func (l *CKADHelmCustomValuesLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADHelmCustomValuesLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADHelmCustomValuesLab) Cert() Cert             { return CertCKAD }
func (l *CKADHelmCustomValuesLab) DomainWeight() int      { return 20 }
func (l *CKADHelmCustomValuesLab) EstimatedTime() int     { return 20 }
func (l *CKADHelmCustomValuesLab) Tags() []string {
	return []string{"helm", "values", "customization"}
}

func (l *CKADHelmCustomValuesLab) Description() string {
	return `A Helm chart is deployed but needs custom configuration. Update the release
with custom values to change the service type to NodePort and set
replica count to 2.

Your task: Upgrade the Helm release with custom values.`
}

func (l *CKADHelmCustomValuesLab) Hints() []string {
	return []string{
		"Use helm upgrade with --set flags or -f values.yaml",
		"Check current values with helm get values",
		"Use --reuse-values to keep existing values",
	}
}

func (l *CKADHelmCustomValuesLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADHelmCustomValuesLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADHelmCustomValuesLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "my-release-nginx",
		"-o", "jsonpath={.spec.type}")
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}
	if strings.TrimSpace(output) != "NodePort" {
		return fmt.Errorf("service type is not NodePort (current: %s)", output)
	}
	return nil
}

func (l *CKADHelmCustomValuesLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current values", Command: "helm get values my-release"},
		{Description: "Upgrade with custom values", Command: "helm upgrade my-release nginx --set service.type=NodePort --set replicaCount=2"},
		{Description: "Verify changes", Command: "helm get values my-release"},
	}
}
