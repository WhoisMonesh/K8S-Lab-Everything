package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&StorageClassWrongProvisioner{})
}

type StorageClassWrongProvisioner struct {
	BaseLab
}

func (l *StorageClassWrongProvisioner) ID() string            { return "storageclass_wrong_provisioner" }
func (l *StorageClassWrongProvisioner) Title() string         { return "StorageClass Wrong Provisioner" }
func (l *StorageClassWrongProvisioner) Category() Category    { return CategoryStorage }
func (l *StorageClassWrongProvisioner) Difficulty() Difficulty { return DifficultyMedium }
func (l *StorageClassWrongProvisioner) EstimatedTime() int    { return 20 }
func (l *StorageClassWrongProvisioner) Tags() []string        { return []string{"storage", "storageclass", "pv"} }

func (l *StorageClassWrongProvisioner) Description() string {
	return `A StorageClass is configured with the wrong provisioner. PVCs using this class are stuck in Pending.
Fix the StorageClass provisioner to match the available CSI driver.`
}

func (l *StorageClassWrongProvisioner) Hints() []string {
	return []string{
		"Check available CSI drivers",
		"Look at the StorageClass provisioner field",
		"Compare with available provisioners in the cluster",
	}
}

func (l *StorageClassWrongProvisioner) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *StorageClassWrongProvisioner) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-storage
provisioner: k8s.io/aws-ebs
parameters:
  type: gp2
volumeBindingMode: Immediate
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: fast-storage`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *StorageClassWrongProvisioner) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pvc", "data-pvc",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return err
	}
	if output == "Pending" {
		return fmt.Errorf("PVC still in Pending state")
	}
	return nil
}

func (l *StorageClassWrongProvisioner) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check StorageClass", Command: "kubectl get storageclass"},
		{Description: "Check available provisioners", Command: "kubectl get csidrivers"},
		{Description: "Fix provisioner", Command: "kubectl patch storageclass fast-storage -p '{\"provisioner\":\"kubernetes.io/no-provisioner\"}'"},
	}
}
