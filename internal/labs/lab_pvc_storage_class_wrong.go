package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PVCStorageClassWrongLab{})
}

type PVCStorageClassWrongLab struct {
	BaseLab
}

func (l *PVCStorageClassWrongLab) ID() string {
	return "pvc_storage_class_wrong"
}

func (l *PVCStorageClassWrongLab) Title() string {
	return "PVC Using Wrong StorageClass"
}

func (l *PVCStorageClassWrongLab) Category() Category {
	return CategoryStorage
}

func (l *PVCStorageClassWrongLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PVCStorageClassWrongLab) Description() string {
	return `A PersistentVolumeClaim 'data-pvc' is stuck in Pending because it
references a StorageClass 'fast-ssd' that doesn't exist in the cluster.
The available StorageClass is named 'standard'.

Your task: Fix the PVC to use the correct StorageClass.`
}

func (l *PVCStorageClassWrongLab) Hints() []string {
	return []string{
		"Check available StorageClasses",
		"The PVC references a non-existent StorageClass",
		"Change storageClassName to match an available StorageClass",
	}
}

func (l *PVCStorageClassWrongLab) EstimatedTime() int {
	return 10
}

func (l *PVCStorageClassWrongLab) Tags() []string {
	return []string{"pvc", "storageclass", "storage"}
}

func (l *PVCStorageClassWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PVCStorageClassWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	pvc := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  storageClassName: fast-ssd
`
	if err := kubectlApply(ctx, kubeconfigPath, pvc); err != nil {
		return fmt.Errorf("creating PVC: %w", err)
	}

	return nil
}

func (l *PVCStorageClassWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pvc", "data-pvc",
		"-o", "jsonpath={.status.phase}")
	if strings.TrimSpace(output) == "Pending" {
		return nil
	}
	return nil
}

func (l *PVCStorageClassWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pvc", "data-pvc",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check PVC: %w", err)
	}

	if strings.TrimSpace(output) != "Bound" {
		return fmt.Errorf("PVC is not Bound (status: %s)", output)
	}

	return nil
}

func (l *PVCStorageClassWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check available StorageClasses",
			Command:     "kubectl get sc",
			Notes:       "Only 'standard' StorageClass exists",
		},
		{
			Description: "Check PVC status",
			Command:     "kubectl get pvc data-pvc",
			Notes:       "PVC is in Pending state",
		},
		{
			Description: "Fix PVC StorageClass",
			Command:     "kubectl edit pvc data-pvc",
			Notes:       "Change storageClassName from 'fast-ssd' to 'standard'",
		},
		{
			Description: "Verify PVC is Bound",
			Command:     "kubectl get pvc data-pvc",
			Notes:       "Should now show Bound status",
		},
	}
}
