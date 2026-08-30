package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&VolumeSnapshotRestoreLab{})
}

type VolumeSnapshotRestoreLab struct {
	BaseLab
}

func (l *VolumeSnapshotRestoreLab) ID() string {
	return "volume_snapshot_restore"
}

func (l *VolumeSnapshotRestoreLab) Title() string {
	return "VolumeSnapshot and Restore Workflow"
}

func (l *VolumeSnapshotRestoreLab) Category() Category {
	return CategoryStorage
}

func (l *VolumeSnapshotRestoreLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *VolumeSnapshotRestoreLab) Description() string {
	return `A PersistentVolumeClaim has important data. A VolumeSnapshot was
taken but the restore process is failing because the
VolumeSnapshotClass is missing.

Your task: Create the missing VolumeSnapshotClass and restore the
data from the snapshot to a new PVC.`
}

func (l *VolumeSnapshotRestoreLab) Hints() []string {
	return []string{
		"Check if a VolumeSnapshotClass exists",
		"The snapshot needs a matching VolumeSnapshotClass",
		"Create the class with the correct driver",
	}
}

func (l *VolumeSnapshotRestoreLab) EstimatedTime() int {
	return 25
}

func (l *VolumeSnapshotRestoreLab) Tags() []string {
	return []string{"volume-snapshot", "restore", "pvc", "storage", "backup"}
}

func (l *VolumeSnapshotRestoreLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *VolumeSnapshotRestoreLab) Break(ctx context.Context, kubeconfigPath string) error {
	pvc := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-pvc
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
`
	if err := kubectlApply(ctx, kubeconfigPath, pvc); err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: data-writer
spec:
  containers:
  - name: writer
    image: busybox:1.36
    command: ['sh', '-c', 'echo "important data" > /data/file.txt && while true; do sleep 30; done']
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: data-pvc
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	snapshot := `apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: data-snapshot
spec:
  volumeSnapshotClassName: csi-vsc
  source:
    persistentVolumeClaimName: data-pvc
`
	return kubectlApply(ctx, kubeconfigPath, snapshot)
}

func (l *VolumeSnapshotRestoreLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "volumesnapshot", "data-snapshot",
		"-o", "jsonpath={.status.readyToUse}")
	if err != nil {
		return nil
	}

	if strings.TrimSpace(output) == "true" {
		return fmt.Errorf("snapshot is ready (expected failure)")
	}

	return nil
}

func (l *VolumeSnapshotRestoreLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "volumesnapshot", "data-snapshot",
		"-o", "jsonpath={.status.readyToUse}")
	if err != nil {
		return fmt.Errorf("checking snapshot: %w", err)
	}

	if strings.TrimSpace(output) != "true" {
		return fmt.Errorf("snapshot not ready (readyToUse: %s)", strings.TrimSpace(output))
	}

	return nil
}

func (l *VolumeSnapshotRestoreLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check snapshot status",
			Command:     "kubectl get volumesnapshot data-snapshot",
			Notes:       "READYTOUSE may be false",
		},
		{
			Description: "Check for VolumeSnapshotClass",
			Command:     "kubectl get volumesnapshotclass",
			Notes:       "Class may be missing",
		},
		{
			Description: "Fix: Create VolumeSnapshotClass",
			Command: `kubectl apply -f - <<EOF
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: csi-vsc
driver: local-path
deletionPolicy: Retain
EOF`,
			Notes: "Create the missing snapshot class",
		},
		{
			Description: "Verify snapshot is ready",
			Command:     "kubectl get volumesnapshot data-snapshot",
			Notes:       "READYTOUSE should be true",
		},
	}
}
