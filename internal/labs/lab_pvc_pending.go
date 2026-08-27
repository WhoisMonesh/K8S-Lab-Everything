package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&PVCPendingLab{})
}

type PVCPendingLab struct {
	BaseLab
}

func (l *PVCPendingLab) ID() string {
	return "pvc_pending"
}

func (l *PVCPendingLab) Title() string {
	return "PersistentVolumeClaim Stuck in Pending"
}

func (l *PVCPendingLab) Category() Category {
	return CategoryStorage
}

func (l *PVCPendingLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PVCPendingLab) Description() string {
	return `A PersistentVolumeClaim named 'app-data' is stuck in Pending state.
Pods trying to use this PVC cannot start.

Your task: Fix the issue so the PVC can bind to a PersistentVolume.`
}

func (l *PVCPendingLab) Hints() []string {
	return []string{
		"Check the PVC status and events",
		"Look for available PersistentVolumes",
		"Compare the StorageClass, capacity, and access modes",
		"The PVC selector might not match any available PV",
	}
}

func (l *PVCPendingLab) EstimatedTime() int {
	return 20
}

func (l *PVCPendingLab) Tags() []string {
	return []string{"storage", "pv", "pvc", "persistent-volume", "troubleshooting"}
}

func (l *PVCPendingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PVCPendingLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create a PersistentVolume with a specific label
	pv := `apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-pv
  labels:
    type: local
spec:
  capacity:
    storage: 5Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/mnt/data"
  storageClassName: manual
`
	if err := kubectlApply(ctx, kubeconfigPath, pv); err != nil {
		return fmt.Errorf("creating PV: %w", err)
	}

	// Create a PVC with mismatched label selector (will not bind)
	pvc := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: app-data
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 3Gi
  storageClassName: manual
  selector:
    matchLabels:
      type: wrong-label
`
	if err := kubectlApply(ctx, kubeconfigPath, pvc); err != nil {
		return fmt.Errorf("creating PVC: %w", err)
	}

	return nil
}

func (l *PVCPendingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Wait for resources to be created
	time.Sleep(10 * time.Second)

	// Check PVC status (should be Pending)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pvc", "app-data", "-o", "jsonpath={.status.phase}")
	_ = output // Should be "Pending"

	return nil
}

func (l *PVCPendingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if the PVC is bound
	output, err := kubectl(ctx, kubeconfigPath, "get", "pvc", "app-data",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check PVC: %w", err)
	}

	if output != "Bound" {
		return fmt.Errorf("PVC is not bound yet (current status: %s)", output)
	}

	// Also check that the PV is bound to this PVC
	output, err = kubectl(ctx, kubeconfigPath, "get", "pvc", "app-data",
		"-o", "jsonpath={.spec.volumeName}")
	if err != nil {
		return fmt.Errorf("failed to check PVC volume name: %w", err)
	}

	if output == "" {
		return fmt.Errorf("PVC does not have a bound volume")
	}

	return nil
}

func (l *PVCPendingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check the PVC status",
			Command:     "kubectl get pvc app-data",
			Notes:       "The PVC should be in 'Pending' status",
		},
		{
			Description: "Describe the PVC to see events",
			Command:     "kubectl describe pvc app-data",
			Notes:       "Look for messages about why it's not binding",
		},
		{
			Description: "List available PersistentVolumes",
			Command:     "kubectl get pv",
			Notes:       "Check if there are any PVs in 'Available' status",
		},
		{
			Description: "Examine the PV labels",
			Command:     "kubectl get pv local-pv -o yaml | grep -A 5 labels",
			Notes:       "The PV has label 'type: local'",
		},
		{
			Description: "Examine the PVC selector",
			Command:     "kubectl get pvc app-data -o yaml | grep -A 5 selector",
			Notes:       "The PVC is looking for 'type: wrong-label', which doesn't match the PV",
		},
		{
			Description: "Fix the PVC selector",
			Command:     "kubectl edit pvc app-data",
			Notes:       "Change the selector from 'type: wrong-label' to 'type: local'",
		},
		{
			Description: "Verify the PVC is now bound",
			Command:     "kubectl get pvc app-data",
			Notes:       "The status should change to 'Bound'",
		},
		{
			Description: "Alternative fix: Remove selector entirely",
			Notes:       "You could also remove the selector section to let Kubernetes auto-bind based on storage class and capacity",
		},
	}
}
