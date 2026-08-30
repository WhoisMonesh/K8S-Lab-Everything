package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&PVCBoundPendingLab{})
}

type PVCBoundPendingLab struct {
	BaseLab
}

func (l *PVCBoundPendingLab) ID() string { return "cka_pvc_bound_pending" }
func (l *PVCBoundPendingLab) Title() string {
	return "Debug PVC Stuck in Pending"
}
func (l *PVCBoundPendingLab) Category() Category     { return CategoryTroubleshooting }
func (l *PVCBoundPendingLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PVCBoundPendingLab) EstimatedTime() int     { return 20 }
func (l *PVCBoundPendingLab) Tags() []string {
	return []string{"pvc", "pending", "storage", "troubleshooting"}
}
func (l *PVCBoundPendingLab) Cert() Cert        { return CertCKA }
func (l *PVCBoundPendingLab) DomainWeight() int { return 30 }

func (l *PVCBoundPendingLab) Description() string {
	return `A PVC is stuck in Pending state. Investigate why the PVC cannot be
bound to a PersistentVolume. Check storage class, access modes, and
capacity requirements.`
}

func (l *PVCBoundPendingLab) Hints() []string {
	return []string{
		"Check PVC events",
		"Verify StorageClass exists",
		"Check PV availability and access modes",
	}
}

func (l *PVCBoundPendingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PVCBoundPendingLab) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: pvc-pending-ns
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: data-pv
spec:
  capacity:
    storage: 1Gi
  volumeMode: Filesystem
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Delete
  storageClassName: manual
  nfs:
    server: nfs-server
    path: /exports
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
  namespace: pvc-pending-ns
spec:
  accessModes:
  - ReadWriteOnce
  storageClassName: slow
  resources:
    requests:
      storage: 1Gi
`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PVCBoundPendingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PVCBoundPendingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pvc", "data-pvc",
		"-n", "pvc-pending-ns", "-o", "jsonpath={.status.phase}")
	if err != nil {
		return err
	}
	if output == "Pending" {
		return fmt.Errorf("PVC still pending")
	}
	return nil
}

func (l *PVCBoundPendingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check PVC", Command: "kubectl get pvc -n pvc-pending-ns"},
		{Description: "Check events", Command: "kubectl describe pvc data-pvc -n pvc-pending-ns"},
		{Description: "Check StorageClass", Command: "kubectl get storageclass"},
		{Description: "Fix issue", Command: "Create matching PV or fix StorageClass"},
	}
}
