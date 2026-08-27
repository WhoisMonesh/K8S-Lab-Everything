package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&VolumeExpansionLab{})
}

type VolumeExpansionLab struct {
	BaseLab
}

func (l *VolumeExpansionLab) ID() string             { return "cka_volume_expansion" }
func (l *VolumeExpansionLab) Title() string          { return "Enable Volume Expansion" }
func (l *VolumeExpansionLab) Category() Category     { return CategoryStorage }
func (l *VolumeExpansionLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *VolumeExpansionLab) EstimatedTime() int     { return 20 }
func (l *VolumeExpansionLab) Tags() []string {
	return []string{"volume-expansion", "storage", "pvc"}
}
func (l *VolumeExpansionLab) Cert() Cert        { return CertCKA }
func (l *VolumeExpansionLab) DomainWeight() int { return 10 }

func (l *VolumeExpansionLab) Description() string {
	return `A PVC needs to be expanded but volume expansion is not enabled on
the StorageClass. Enable allowVolumeExpansion and resize the PVC.`
}

func (l *VolumeExpansionLab) Hints() []string {
	return []string{
		"Check the StorageClass allowVolumeExpansion setting",
		"Patch the StorageClass to enable expansion",
		"Patch the PVC to request more storage",
	}
}

func (l *VolumeExpansionLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *VolumeExpansionLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *VolumeExpansionLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pvc", "expand-pvc",
		"-n", "expand-ns", "-o", "jsonpath={.spec.resources.requests.storage}")
	if err != nil {
		return err
	}
	if output == "1Gi" {
		return fmt.Errorf("PVC not expanded")
	}
	return nil
}

func (l *VolumeExpansionLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check StorageClass", Command: "kubectl get storageclass -o yaml"},
		{Description: "Enable expansion", Command: "kubectl patch storageclass standard -p '{\"allowVolumeExpansion\":true}'"},
		{Description: "Expand PVC", Command: "kubectl patch pvc expand-pvc -n expand-ns -p '{\"spec\":{\"resources\":{\"requests\":{\"storage\":\"5Gi\"}}}}'"},
		{Description: "Verify", Command: "kubectl get pvc expand-pvc -n expand-ns"},
	}
}
