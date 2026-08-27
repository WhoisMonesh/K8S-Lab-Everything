package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&PVCAccessModesLab{})
}

type PVCAccessModesLab struct {
	BaseLab
}

func (l *PVCAccessModesLab) ID() string             { return "cka_pvc_access_modes" }
func (l *PVCAccessModesLab) Title() string          { return "Understand PVC Access Mode Conflicts" }
func (l *PVCAccessModesLab) Category() Category     { return CategoryStorage }
func (l *PVCAccessModesLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PVCAccessModesLab) EstimatedTime() int     { return 20 }
func (l *PVCAccessModesLab) Tags() []string {
	return []string{"pvc", "access-modes", "storage"}
}
func (l *PVCAccessModesLab) Cert() Cert        { return CertCKA }
func (l *PVCAccessModesLab) DomainWeight() int { return 10 }

func (l *PVCAccessModesLab) Description() string {
	return `A PVC is stuck in Pending because it requests ReadWriteMany access
mode but the available PV only supports ReadWriteOnce. Fix the PVC to
use the correct access mode.`
}

func (l *PVCAccessModesLab) Hints() []string {
	return []string{
		"Check PVC access modes",
		"Check PV access modes",
		"Match PVC access mode to PV capabilities",
	}
}

func (l *PVCAccessModesLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PVCAccessModesLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *PVCAccessModesLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pvc", "data-pvc",
		"-n", "pvc-ns", "-o", "jsonpath={.spec.accessModes}")
	if err != nil {
		return err
	}
	if strings.Contains(output, "ReadWriteMany") {
		return fmt.Errorf("PVC still requesting ReadWriteMany")
	}
	return nil
}

func (l *PVCAccessModesLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check PVC", Command: "kubectl get pvc data-pvc -n pvc-ns -o yaml"},
		{Description: "Check PV", Command: "kubectl get pv -o wide"},
		{Description: "Fix PVC access mode", Command: "kubectl patch pvc data-pvc -n pvc-ns -p '{\"spec\":{\"accessModes\":[\"ReadWriteOnce\"]}}'"},
		{Description: "Verify bound", Command: "kubectl get pvc data-pvc -n pvc-ns"},
	}
}
