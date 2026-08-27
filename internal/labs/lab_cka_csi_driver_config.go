package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&CSIDriverConfigLab{})
}

type CSIDriverConfigLab struct {
	BaseLab
}

func (l *CSIDriverConfigLab) ID() string             { return "cka_csi_driver_config" }
func (l *CSIDriverConfigLab) Title() string          { return "Configure CSI Driver" }
func (l *CSIDriverConfigLab) Category() Category     { return CategoryStorage }
func (l *CSIDriverConfigLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CSIDriverConfigLab) EstimatedTime() int     { return 25 }
func (l *CSIDriverConfigLab) Tags() []string {
	return []string{"csi", "driver", "storage", "volume"}
}
func (l *CSIDriverConfigLab) Cert() Cert        { return CertCKA }
func (l *CSIDriverConfigLab) DomainWeight() int { return 10 }

func (l *CSIDriverConfigLab) Description() string {
	return `A CSI driver deployment is missing the required CSIDriver resource.
Create the CSIDriver resource and configure the driver to support
persistent and ephemeral volumes.`
}

func (l *CSIDriverConfigLab) Hints() []string {
	return []string{
		"Create a CSIDriver resource",
		"Set attachRequired to true",
		"Configure podInfoOnMount for better volume management",
	}
}

func (l *CSIDriverConfigLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CSIDriverConfigLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CSIDriverConfigLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "csidriver", "my-csi-driver",
		"-o", "jsonpath={.spec.attachRequired}")
	if err != nil {
		return err
	}
	if output != "true" {
		return fmt.Errorf("CSIDriver not properly configured")
	}
	return nil
}

func (l *CSIDriverConfigLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check CSIDrivers", Command: "kubectl get csidriver"},
		{Description: "Create CSIDriver", Command: "cat <<EOF | kubectl apply -f -\napiVersion: storage.k8s.io/v1\nkind: CSIDriver\nmetadata:\n  name: my-csi-driver\nspec:\n  attachRequired: true\n  podInfoOnMount: true\n  fsGroupPolicy: File\nEOF"},
		{Description: "Verify", Command: "kubectl get csidriver my-csi-driver -o yaml"},
	}
}
