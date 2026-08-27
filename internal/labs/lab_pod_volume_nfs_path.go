package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodVolumeNFSPathLab{})
}

type PodVolumeNFSPathLab struct {
	BaseLab
}

func (l *PodVolumeNFSPathLab) ID() string {
	return "pod_volume_nfs_path"
}

func (l *PodVolumeNFSPathLab) Title() string {
	return "NFS Volume Path Wrong"
}

func (l *PodVolumeNFSPathLab) Category() Category {
	return CategoryStorage
}

func (l *PodVolumeNFSPathLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodVolumeNFSPathLab) Description() string {
	return `A pod 'nfs-reader' is failing because it mounts an NFS volume at the
wrong path. The NFS export is at /exports/data but the pod tries to
mount /exports/app.

Your task: Fix the NFS volume path in the pod spec.`
}

func (l *PodVolumeNFSPathLab) Hints() []string {
	return []string{
		"Check the pod's volume configuration",
		"The path field specifies the NFS export path",
		"Verify the correct path with the NFS server",
	}
}

func (l *PodVolumeNFSPathLab) EstimatedTime() int {
	return 10
}

func (l *PodVolumeNFSPathLab) Tags() []string {
	return []string{"nfs", "volume", "path", "storage"}
}

func (l *PodVolumeNFSPathLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodVolumeNFSPathLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: nfs-reader
  namespace: default
spec:
  containers:
  - name: reader
    image: busybox:1.36
    command: ['sh', '-c', 'ls /mnt/data && sleep 3600']
    volumeMounts:
    - name: nfs-vol
      mountPath: /mnt/data
  volumes:
  - name: nfs-vol
    nfs:
      server: 10.0.0.10
      path: /exports/app
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodVolumeNFSPathLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodVolumeNFSPathLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "nfs-reader",
		"-o", "jsonpath={.spec.volumes[0].nfs.path}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) == "/exports/app" {
		return fmt.Errorf("NFS path is still /exports/app")
	}

	return nil
}

func (l *PodVolumeNFSPathLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check NFS volume configuration",
			Command:     "kubectl get pod nfs-reader -o yaml | grep -A 5 nfs",
			Notes:       "path is /exports/app instead of /exports/data",
		},
		{
			Description: "Fix NFS path",
			Command:     "kubectl edit pod nfs-reader",
			Notes:       "Change path from /exports/app to /exports/data",
		},
		{
			Description: "Verify volume mount",
			Command:     "kubectl get pod nfs-reader -o yaml | grep -A 5 nfs",
			Notes:       "path should now be /exports/data",
		},
	}
}
