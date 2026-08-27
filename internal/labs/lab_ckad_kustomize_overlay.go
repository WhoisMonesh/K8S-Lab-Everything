package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADKustomizeOverlayLab{})
}

type CKADKustomizeOverlayLab struct {
	BaseLab
}

func (l *CKADKustomizeOverlayLab) ID() string             { return "ckad_kustomize_overlay" }
func (l *CKADKustomizeOverlayLab) Title() string          { return "Create Kustomize Overlay" }
func (l *CKADKustomizeOverlayLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADKustomizeOverlayLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKADKustomizeOverlayLab) Cert() Cert             { return CertCKAD }
func (l *CKADKustomizeOverlayLab) DomainWeight() int      { return 20 }
func (l *CKADKustomizeOverlayLab) EstimatedTime() int     { return 25 }
func (l *CKADKustomizeOverlayLab) Tags() []string {
	return []string{"kustomize", "overlay", "env-specific"}
}

func (l *CKADKustomizeOverlayLab) Description() string {
	return `Create a Kustomize overlay for the production environment that modifies
the base configuration to use 3 replicas and adds a resource limit.

Your task: Create an overlay that references the base and applies patches.`
}

func (l *CKADKustomizeOverlayLab) Hints() []string {
	return []string{
		"Overlays reference the base using resources or bases field",
		"Use patches or strategic merge patches to modify resources",
		"Namespace can be set in the overlay",
	}
}

func (l *CKADKustomizeOverlayLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADKustomizeOverlayLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADKustomizeOverlayLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp",
		"-n", "production", "-o", "jsonpath={.spec.replicas}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if strings.TrimSpace(output) != "3" {
		return fmt.Errorf("replicas not set to 3 (current: %s)", output)
	}
	return nil
}

func (l *CKADKustomizeOverlayLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create overlay directory", Command: "mkdir -p overlay/production"},
		{Description: "Create kustomization.yaml", Command: "Reference base and add namespace: production"},
		{Description: "Create patch file", Command: "Patch deployment to set replicas: 3 and add resource limits"},
		{Description: "Apply overlay", Command: "kubectl apply -k overlay/production/"},
	}
}
