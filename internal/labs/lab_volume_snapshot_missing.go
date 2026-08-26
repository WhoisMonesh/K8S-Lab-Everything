package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&VolumeSnapshotMissing{})
}

type VolumeSnapshotMissing struct {
	BaseLab
}

func (l *VolumeSnapshotMissing) ID() string             { return "volume_snapshot_missing" }
func (l *VolumeSnapshotMissing) Title() string          { return "Volume Snapshot Not Found" }
func (l *VolumeSnapshotMissing) Category() Category     { return CategoryStorage }
func (l *VolumeSnapshotMissing) Difficulty() Difficulty { return DifficultyMedium }
func (l *VolumeSnapshotMissing) EstimatedTime() int     { return 20 }
func (l *VolumeSnapshotMissing) Tags() []string         { return []string{"storage", "snapshot", "backup"} }

func (l *VolumeSnapshotMissing) Description() string {
	return `A StatefulSet needs to be restored from a VolumeSnapshot but the snapshot doesn't exist.
Create the required VolumeSnapshot and restore the data.`
}

func (l *VolumeSnapshotMissing) Hints() []string {
	return []string{
		"Check if VolumeSnapshotClass exists",
		"Verify PVC exists to snapshot from",
		"Create the VolumeSnapshot resource",
	}
}

func (l *VolumeSnapshotMissing) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *VolumeSnapshotMissing) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: restore-target
spec:
  replicas: 1
  selector:
    matchLabels:
      app: restore-target
  serviceName: restore-target
  template:
    metadata:
      labels:
        app: restore-target
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        volumeMounts:
        - name: data
          mountPath: /usr/share/nginx/html
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 5Gi`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *VolumeSnapshotMissing) Verify(ctx context.Context, kubeconfigPath string) error {
	_, err := kubectl(ctx, kubeconfigPath, "get", "volumesnapshot", "data-snapshot")
	if err != nil {
		return fmt.Errorf("volume snapshot not found")
	}
	return nil
}

func (l *VolumeSnapshotMissing) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check existing PVCs", Command: "kubectl get pvc"},
		{Description: "Create VolumeSnapshot", Command: "kubectl create -f - <<EOF\napiVersion: snapshot.storage.k8s.io/v1\nkind: VolumeSnapshot\nmetadata:\n  name: data-snapshot\nspec:\n  volumeSnapshotClassName: csi-snapclass\n  source:\n    persistentVolumeClaimName: data-pvc\nEOF"},
		{Description: "Verify snapshot", Command: "kubectl get volumesnapshot"},
	}
}
