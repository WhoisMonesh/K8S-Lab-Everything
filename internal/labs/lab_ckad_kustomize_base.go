package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADKustomizeBaseLab{})
}

type CKADKustomizeBaseLab struct {
	BaseLab
}

func (l *CKADKustomizeBaseLab) ID() string             { return "ckad_kustomize_base" }
func (l *CKADKustomizeBaseLab) Title() string          { return "Create Kustomize Base" }
func (l *CKADKustomizeBaseLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADKustomizeBaseLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADKustomizeBaseLab) Cert() Cert             { return CertCKAD }
func (l *CKADKustomizeBaseLab) DomainWeight() int      { return 20 }
func (l *CKADKustomizeBaseLab) EstimatedTime() int     { return 20 }
func (l *CKADKustomizeBaseLab) Tags() []string {
	return []string{"kustomize", "base", "deployment"}
}

func (l *CKADKustomizeBaseLab) Description() string {
	return `Create a Kustomize base configuration for a web application that includes
a Deployment and Service. The base should be reusable across environments.

Your task: Create a kustomization.yaml file with the base resources.`
}

func (l *CKADKustomizeBaseLab) Hints() []string {
	return []string{
		"Create a directory structure with base/ folder",
		"kustomization.yaml lists the resources",
		"Use commonLabels for consistent labeling",
	}
}

func (l *CKADKustomizeBaseLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADKustomizeBaseLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADKustomizeBaseLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=webapp",
		"-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}
	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pods not running (status: %s)", output)
	}
	return nil
}

func (l *CKADKustomizeBaseLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create base directory", Command: "mkdir -p base"},
		{Description: "Create deployment.yaml", Command: "Create deployment manifest in base/"},
		{Description: "Create service.yaml", Command: "Create service manifest in base/"},
		{Description: "Create kustomization.yaml", Command: "Create kustomization.yaml listing resources and commonLabels"},
		{Description: "Apply base", Command: "kubectl apply -k base/"},
	}
}
