package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PVCCapacityInsufficientLab{})
}

type PVCCapacityInsufficientLab struct {
	BaseLab
}

func (l *PVCCapacityInsufficientLab) ID() string {
	return "pvc_capacity_insufficient"
}

func (l *PVCCapacityInsufficientLab) Title() string {
	return "PVC Capacity Too Small"
}

func (l *PVCCapacityInsufficientLab) Category() Category {
	return CategoryStorage
}

func (l *PVCCapacityInsufficientLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *PVCCapacityInsufficientLab) Description() string {
	return `A PersistentVolumeClaim 'db-data' requests 1Gi but the application
needs at least 5Gi of storage. The PVC is Bound but the pod runs out
of space quickly.

Your task: Increase the PVC capacity to meet application requirements.`
}

func (l *PVCCapacityInsufficientLab) Hints() []string {
	return []string{
		"Check the PVC storage request",
		"1Gi is insufficient for the application",
		"Increase storage request to 5Gi or more",
	}
}

func (l *PVCCapacityInsufficientLab) EstimatedTime() int {
	return 10
}

func (l *PVCCapacityInsufficientLab) Tags() []string {
	return []string{"pvc", "capacity", "storage"}
}

func (l *PVCCapacityInsufficientLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PVCCapacityInsufficientLab) Break(ctx context.Context, kubeconfigPath string) error {
	pvc := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: db-data
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: standard
`
	if err := kubectlApply(ctx, kubeconfigPath, pvc); err != nil {
		return fmt.Errorf("creating PVC: %w", err)
	}

	return nil
}

func (l *PVCCapacityInsufficientLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *PVCCapacityInsufficientLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pvc", "db-data",
		"-o", "jsonpath={.spec.resources.requests.storage}")
	if err != nil {
		return fmt.Errorf("failed to check PVC: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "1Gi" {
		return fmt.Errorf("PVC capacity is still 1Gi")
	}

	return nil
}

func (l *PVCCapacityInsufficientLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check PVC capacity",
			Command:     "kubectl get pvc db-data",
			Notes:       "CURRENT CAPACITY shows 1Gi",
		},
		{
			Description: "Increase PVC capacity",
			Command:     "kubectl patch pvc db-data --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/resources/requests/storage\",\"value\":\"5Gi\"}]'",
			Notes:       "Set storage request to 5Gi",
		},
		{
			Description: "Verify new capacity",
			Command:     "kubectl get pvc db-data",
			Notes:       "Should now show 5Gi capacity",
		},
	}
}
