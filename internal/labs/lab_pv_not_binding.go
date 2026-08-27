package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PVNotBindingLab{})
}

type PVNotBindingLab struct {
	BaseLab
}

func (l *PVNotBindingLab) ID() string {
	return "pv_not_binding"
}

func (l *PVNotBindingLab) Title() string {
	return "PersistentVolumeClaim Stuck in Pending"
}

func (l *PVNotBindingLab) Category() Category {
	return CategoryStorage
}

func (l *PVNotBindingLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PVNotBindingLab) Description() string {
	return `A PersistentVolumeClaim 'data-pvc' is stuck in Pending state.
The PVC cannot bind to any available PersistentVolume.

Your task: Fix the storage configuration so the PVC can bind to a PV.`
}

func (l *PVNotBindingLab) Hints() []string {
	return []string{
		"Check the PVC status",
		"Look at the PVC and PV specifications",
		"Compare the access modes between PVC and PV",
		"Check if the storageClassName matches",
	}
}

func (l *PVNotBindingLab) EstimatedTime() int {
	return 20
}

func (l *PVNotBindingLab) Tags() []string {
	return []string{"storage", "pvc", "pv", "persistent-volume"}
}

func (l *PVNotBindingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PVNotBindingLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create PV with ReadWriteOnce access mode
	pv := `apiVersion: v1
kind: PersistentVolume
metadata:
  name: data-pv
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: slow
  hostPath:
    path: /mnt/data
`
	if err := kubectlApply(ctx, kubeconfigPath, pv); err != nil {
		return fmt.Errorf("creating PV: %w", err)
	}

	// Create PVC with ReadWriteMany (won't bind to ReadWriteOnce PV)
	pvc := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 1Gi
  storageClassName: slow
`
	if err := kubectlApply(ctx, kubeconfigPath, pvc); err != nil {
		return fmt.Errorf("creating PVC: %w", err)
	}

	return nil
}

func (l *PVNotBindingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PVNotBindingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if PVC is Bound
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

func (l *PVNotBindingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check PVC status",
			Command:     "kubectl get pvc data-pvc",
			Notes:       "The PVC is stuck in Pending state",
		},
		{
			Description: "Describe the PVC",
			Command:     "kubectl describe pvc data-pvc",
			Notes:       "Look for events showing no PV matches the claim",
		},
		{
			Description: "Check PV status",
			Command:     "kubectl get pv data-pv",
			Notes:       "The PV exists but is not binding to the PVC",
		},
		{
			Description: "Compare PVC and PV access modes",
			Command:     "kubectl get pv data-pv -o yaml | grep -A 3 accessModes && kubectl get pvc data-pvc -o yaml | grep -A 3 accessModes",
			Notes:       "PV has ReadWriteOnce but PVC requests ReadWriteMany",
		},
		{
			Description: "Fix the PVC access mode",
			Command:     "kubectl edit pvc data-pvc",
			Notes:       "Change accessModes from ReadWriteMany to ReadWriteOnce",
		},
		{
			Description: "Verify PVC is Bound",
			Command:     "kubectl get pvc data-pvc",
			Notes:       "The PVC should now show as Bound",
		},
	}
}
