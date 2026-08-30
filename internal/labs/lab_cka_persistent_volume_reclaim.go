package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&PersistentVolumeReclaimCKALab{})
}

type PersistentVolumeReclaimCKALab struct {
	BaseLab
}

func (l *PersistentVolumeReclaimCKALab) ID() string {
	return "cka_persistent_volume_reclaim"
}
func (l *PersistentVolumeReclaimCKALab) Title() string {
	return "Configure PV Reclaim Policies"
}
func (l *PersistentVolumeReclaimCKALab) Category() Category     { return CategoryStorage }
func (l *PersistentVolumeReclaimCKALab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PersistentVolumeReclaimCKALab) EstimatedTime() int     { return 15 }
func (l *PersistentVolumeReclaimCKALab) Tags() []string {
	return []string{"pv", "reclaim-policy", "storage"}
}
func (l *PersistentVolumeReclaimCKALab) Cert() Cert        { return CertCKA }
func (l *PersistentVolumeReclaimCKALab) DomainWeight() int { return 10 }

func (l *PersistentVolumeReclaimCKALab) Description() string {
	return `A PersistentVolume is set to Retain policy preventing cleanup. Change
the reclaim policy to Delete so the volume is automatically cleaned up
when the PVC is deleted.`
}

func (l *PersistentVolumeReclaimCKALab) Hints() []string {
	return []string{
		"Check the PV reclaim policy",
		"Patch the PV to change the policy",
		"Delete the PVC to test",
	}
}

func (l *PersistentVolumeReclaimCKALab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PersistentVolumeReclaimCKALab) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: PersistentVolume
metadata:
  name: data-pv
spec:
  capacity:
    storage: 1Gi
  volumeMode: Filesystem
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: manual
  nfs:
    server: nfs-server
    path: /exports
`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PersistentVolumeReclaimCKALab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PersistentVolumeReclaimCKALab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pv", "data-pv",
		"-o", "jsonpath={.spec.persistentVolumeReclaimPolicy}")
	if err != nil {
		return err
	}
	if output == "Retain" {
		return fmt.Errorf("PV still has Retain policy")
	}
	return nil
}

func (l *PersistentVolumeReclaimCKALab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check PV", Command: "kubectl get pv data-pv -o yaml"},
		{Description: "Patch PV", Command: "kubectl patch pv data-pv -p '{\"spec\":{\"persistentVolumeReclaimPolicy\":\"Delete\"}}'"},
		{Description: "Verify", Command: "kubectl get pv data-pv -o jsonpath='{.spec.persistentVolumeReclaimPolicy}'"},
	}
}
