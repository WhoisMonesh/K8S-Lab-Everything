package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADAPIDeprecationLab{})
}

type CKADAPIDeprecationLab struct {
	BaseLab
}

func (l *CKADAPIDeprecationLab) ID() string             { return "ckad_api_deprecation" }
func (l *CKADAPIDeprecationLab) Title() string          { return "Handle API Deprecations" }
func (l *CKADAPIDeprecationLab) Category() Category     { return CategoryAppObservability }
func (l *CKADAPIDeprecationLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADAPIDeprecationLab) Cert() Cert             { return CertCKAD }
func (l *CKADAPIDeprecationLab) DomainWeight() int      { return 15 }
func (l *CKADAPIDeprecationLab) EstimatedTime() int     { return 20 }
func (l *CKADAPIDeprecationLab) Tags() []string {
	return []string{"api-deprecation", "migrations", "versioning"}
}

func (l *CKADAPIDeprecationLab) Description() string {
	return `A manifest uses deprecated API versions that will be removed in future
Kubernetes releases. Update the manifests to use the stable API versions.

Your task: Migrate the deprecated extensions/v1beta1 Deployment to apps/v1.`
}

func (l *CKADAPIDeprecationLab) Hints() []string {
	return []string{
		"extensions/v1beta1 Deployment is deprecated, use apps/v1",
		"apps/v1 requires a selector in the spec",
		"Use kubectl convert or manually update the apiVersion",
	}
}

func (l *CKADAPIDeprecationLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADAPIDeprecationLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADAPIDeprecationLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployments",
		"-o", "jsonpath={.items[*].apiVersion}")
	if err != nil {
		return fmt.Errorf("failed to get deployments: %w", err)
	}
	if strings.Contains(output, "extensions") || strings.Contains(output, "beta") {
		return fmt.Errorf("deprecated API versions still in use")
	}
	return nil
}

func (l *CKADAPIDeprecationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current API versions", Command: "kubectl get deployments -o yaml | grep apiVersion"},
		{Description: "Update apiVersion", Command: "Change extensions/v1beta1 to apps/v1"},
		{Description: "Add selector", Command: "Add required selector field for apps/v1"},
		{Description: "Apply updated manifest", Command: "kubectl apply -f deployment.yaml"},
	}
}
