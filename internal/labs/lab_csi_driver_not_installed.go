package labs

import (
	"context"
)

func init() {
	Register(&CSIDriverNotInstalled{})
}

type CSIDriverNotInstalled struct {
	BaseLab
}

func (l *CSIDriverNotInstalled) ID() string             { return "csi_driver_not_installed" }
func (l *CSIDriverNotInstalled) Title() string          { return "CSI Driver Not Installed" }
func (l *CSIDriverNotInstalled) Category() Category     { return CategoryStorage }
func (l *CSIDriverNotInstalled) Difficulty() Difficulty { return DifficultyHard }
func (l *CSIDriverNotInstalled) EstimatedTime() int     { return 25 }
func (l *CSIDriverNotInstalled) Tags() []string         { return []string{"storage", "csi", "driver"} }

func (l *CSIDriverNotInstalled) Description() string {
	return `A StorageClass references a CSI driver that doesn't exist.
Install the required CSI driver or fix the StorageClass to use an available provisioner.`
}

func (l *CSIDriverNotInstalled) Hints() []string {
	return []string{
		"Check available CSI drivers",
		"Look at the StorageClass provisioner",
		"Install or configure an available driver",
	}
}

func (l *CSIDriverNotInstalled) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CSIDriverNotInstalled) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: csi-storage
provisioner: csi-driver-that-does-not-exist
parameters:
  type: standard
volumeBindingMode: Immediate`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *CSIDriverNotInstalled) Verify(ctx context.Context, kubeconfigPath string) error {
	_, err := kubectl(ctx, kubeconfigPath, "get", "csidrivers")
	if err != nil {
		return err
	}
	return nil
}

func (l *CSIDriverNotInstalled) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check CSI drivers", Command: "kubectl get csidrivers"},
		{Description: "Check StorageClass", Command: "kubectl get storageclass csi-storage -o yaml"},
		{Description: "Fix provisioner", Command: "kubectl patch storageclass csi-storage -p '{\"provisioner\":\"kubernetes.io/no-provisioner\"}'"},
	}
}
