package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PVAccessModeWrongLab{})
}

type PVAccessModeWrongLab struct {
	BaseLab
}

func (l *PVAccessModeWrongLab) ID() string {
	return "pv_access_mode_wrong"
}

func (l *PVAccessModeWrongLab) Title() string {
	return "PV Access Mode Mismatch"
}

func (l *PVAccessModeWrongLab) Category() Category {
	return CategoryStorage
}

func (l *PVAccessModeWrongLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PVAccessModeWrongLab) Description() string {
	return `A PersistentVolumeClaim 'shared-pvc' requests ReadWriteMany access mode
but the available PersistentVolume only supports ReadWriteOnce. The PVC
cannot bind to the PV due to access mode mismatch.

Your task: Fix either the PVC or PV to match access modes.`
}

func (l *PVAccessModeWrongLab) Hints() []string {
	return []string{
		"Check PVC and PV access modes",
		"ReadWriteMany requires a storage backend that supports multiple node access",
		"Either change PVC to ReadWriteOnce or use a different PV",
	}
}

func (l *PVAccessModeWrongLab) EstimatedTime() int {
	return 15
}

func (l *PVAccessModeWrongLab) Tags() []string {
	return []string{"pv", "pvc", "access-mode", "storage"}
}

func (l *PVAccessModeWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PVAccessModeWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	pv := `apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: standard
  hostPath:
    path: /mnt/data
`
	if err := kubectlApply(ctx, kubeconfigPath, pv); err != nil {
		return fmt.Errorf("creating PV: %w", err)
	}

	pvc := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: shared-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteMany
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

func (l *PVAccessModeWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pvc", "shared-pvc",
		"-o", "jsonpath={.status.phase}")
	if strings.TrimSpace(output) == "Pending" {
		return nil
	}
	return nil
}

func (l *PVAccessModeWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pvc", "shared-pvc",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check PVC: %w", err)
	}

	if strings.TrimSpace(output) != "Bound" {
		return fmt.Errorf("PVC is not Bound (status: %s)", output)
	}

	return nil
}

func (l *PVAccessModeWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check PVC access mode",
			Command:     "kubectl get pvc shared-pvc -o yaml | grep accessModes",
			Notes:       "PVC requests ReadWriteMany",
		},
		{
			Description: "Check PV access mode",
			Command:     "kubectl get pv local-pv -o yaml | grep accessModes",
			Notes:       "PV only supports ReadWriteOnce",
		},
		{
			Description: "Fix PVC access mode",
			Command:     "kubectl edit pvc shared-pvc",
			Notes:       "Change accessModes from ReadWriteMany to ReadWriteOnce",
		},
		{
			Description: "Verify PVC is Bound",
			Command:     "kubectl get pvc shared-pvc",
			Notes:       "Should now show Bound status",
		},
	}
}
