package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PVCClaimRefWrongLab{})
}

type PVCClaimRefWrongLab struct {
	BaseLab
}

func (l *PVCClaimRefWrongLab) ID() string {
	return "pvc_claim_ref_wrong"
}

func (l *PVCClaimRefWrongLab) Title() string {
	return "PVC claimRef Mismatch"
}

func (l *PVCClaimRefWrongLab) Category() Category {
	return CategoryStorage
}

func (l *PVCClaimRefWrongLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *PVCClaimRefWrongLab) Description() string {
	return `A PersistentVolume 'reserved-pv' has a claimRef pointing to a PVC
'data-pvc' that doesn't exist. This prevents the PV from binding to
any other PVC, including the intended 'app-data-pvc'.

Your task: Fix the PV claimRef to allow proper binding.`
}

func (l *PVCClaimRefWrongLab) Hints() []string {
	return []string{
		"Check the PV claimRef",
		"A claimRef to a non-existent PVC prevents binding",
		"Remove or update the claimRef to match the actual PVC",
	}
}

func (l *PVCClaimRefWrongLab) EstimatedTime() int {
	return 15
}

func (l *PVCClaimRefWrongLab) Tags() []string {
	return []string{"pv", "pvc", "claimref", "storage"}
}

func (l *PVCClaimRefWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PVCClaimRefWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	pv := `apiVersion: v1
kind: PersistentVolume
metadata:
  name: reserved-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: standard
  claimRef:
    namespace: default
    name: data-pvc
  hostPath:
    path: /mnt/data
`
	if err := kubectlApply(ctx, kubeconfigPath, pv); err != nil {
		return fmt.Errorf("creating PV: %w", err)
	}

	pvc := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: app-data-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: standard
`
	if err := kubectlApply(ctx, kubeconfigPath, pvc); err != nil {
		return fmt.Errorf("creating PVC: %w", err)
	}

	return nil
}

func (l *PVCClaimRefWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pvc", "app-data-pvc",
		"-o", "jsonpath={.status.phase}")
	if strings.TrimSpace(output) == "Pending" {
		return nil
	}
	return nil
}

func (l *PVCClaimRefWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pvc", "app-data-pvc",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check PVC: %w", err)
	}

	if strings.TrimSpace(output) != "Bound" {
		return fmt.Errorf("PVC is not Bound (status: %s)", output)
	}

	return nil
}

func (l *PVCClaimRefWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check PV claimRef",
			Command:     "kubectl get pv reserved-pv -o yaml | grep -A 5 claimRef",
			Notes:       "claimRef points to data-pvc which doesn't exist",
		},
		{
			Description: "Fix claimRef",
			Command:     "kubectl edit pv reserved-pv",
			Notes:       "Change claimRef name from data-pvc to app-data-pvc",
		},
		{
			Description: "Verify PVC is Bound",
			Command:     "kubectl get pvc app-data-pvc",
			Notes:       "Should now show Bound status",
		},
	}
}
