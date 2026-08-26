package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PersistentVolumeReclaimPolicy{})
}

type PersistentVolumeReclaimPolicy struct {
	BaseLab
}

func (l *PersistentVolumeReclaimPolicy) ID() string            { return "persistent_volume_reclaim_policy" }
func (l *PersistentVolumeReclaimPolicy) Title() string         { return "PV Reclaim Policy Prevents PVC Delete" }
func (l *PersistentVolumeReclaimPolicy) Category() Category    { return CategoryStorage }
func (l *PersistentVolumeReclaimPolicy) Difficulty() Difficulty { return DifficultyMedium }
func (l *PersistentVolumeReclaimPolicy) EstimatedTime() int    { return 15 }
func (l *PersistentVolumeReclaimPolicy) Tags() []string        { return []string{"storage", "pv", "reclaim"} }

func (l *PersistentVolumeReclaimPolicy) Description() string {
	return `A PVC cannot be deleted because the backing PV has a reclaim policy that prevents deletion.
Change the reclaim policy to allow the PVC to be deleted.`
}

func (l *PersistentVolumeReclaimPolicy) Hints() []string {
	return []string{
		"Check the PV reclaim policy",
		"Look at the PVC finalizers",
		"Patch the PV to change reclaim policy",
	}
}

func (l *PersistentVolumeReclaimPolicy) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PersistentVolumeReclaimPolicy) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: PersistentVolume
metadata:
  name: retained-pv
spec:
  capacity:
    storage: 5Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  hostPath:
    path: /tmp/data
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: retained-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
  volumeName: retained-pv`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PersistentVolumeReclaimPolicy) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pv", "retained-pv",
		"-o", "jsonpath={.spec.persistentVolumeReclaimPolicy}")
	if err != nil {
		return err
	}
	if output == "Retain" {
		return fmt.Errorf("PV still has Retain policy")
	}
	return nil
}

func (l *PersistentVolumeReclaimPolicy) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check PV reclaim policy", Command: "kubectl get pv retained-pv -o jsonpath='{.spec.persistentVolumeReclaimPolicy}'"},
		{Description: "Patch PV", Command: "kubectl patch pv retained-pv -p '{\"spec\":{\"persistentVolumeReclaimPolicy\":\"Delete\"}}'"},
		{Description: "Delete PVC", Command: "kubectl delete pvc retained-pvc"},
	}
}
