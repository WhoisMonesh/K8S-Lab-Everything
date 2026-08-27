package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADVolumeNFSMountLab{})
}

type CKADVolumeNFSMountLab struct {
	BaseLab
}

func (l *CKADVolumeNFSMountLab) ID() string             { return "ckad_volume_nfs_mount" }
func (l *CKADVolumeNFSMountLab) Title() string          { return "Mount NFS Shared Volume" }
func (l *CKADVolumeNFSMountLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADVolumeNFSMountLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKADVolumeNFSMountLab) Cert() Cert             { return CertCKAD }
func (l *CKADVolumeNFSMountLab) DomainWeight() int      { return 20 }
func (l *CKADVolumeNFSMountLab) EstimatedTime() int     { return 25 }
func (l *CKADVolumeNFSMountLab) Tags() []string {
	return []string{"nfs", "volumes", "shared-storage"}
}

func (l *CKADVolumeNFSMountLab) Description() string {
	return `Multiple pods need to share data through an NFS volume.
An NFS server is available at nfs-server.local:/exports/data.

Your task: Configure two pods to mount the same NFS volume for shared access.`
}

func (l *CKADVolumeNFSMountLab) Hints() []string {
	return []string{
		"Use volumes with nfs type",
		"Both pods should mount the same NFS path",
		"Set readOnly: false for read-write access",
	}
}

func (l *CKADVolumeNFSMountLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADVolumeNFSMountLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod1 := `apiVersion: v1
kind: Pod
metadata:
  name: nfs-writer
  labels:
    app: nfs-writer
spec:
  containers:
  - name: writer
    image: busybox:1.36
    command: ['sh', '-c', 'while true; do sleep 3600; done']`
	if err := kubectlApply(ctx, kubeconfigPath, pod1); err != nil {
		return fmt.Errorf("creating pod1: %w", err)
	}

	pod2 := `apiVersion: v1
kind: Pod
metadata:
  name: nfs-reader
  labels:
    app: nfs-reader
spec:
  containers:
  - name: reader
    image: busybox:1.36
    command: ['sh', '-c', 'while true; do sleep 3600; done']`
	return kubectlApply(ctx, kubeconfigPath, pod2)
}

func (l *CKADVolumeNFSMountLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output1, err := kubectl(ctx, kubeconfigPath, "get", "pod", "nfs-writer",
		"-o", "jsonpath={.spec.volumes[*].nfs.server}")
	if err != nil {
		return fmt.Errorf("failed to get pod1: %w", err)
	}
	if strings.TrimSpace(output1) == "" {
		return fmt.Errorf("no NFS volume found on nfs-writer")
	}

	output2, err := kubectl(ctx, kubeconfigPath, "get", "pod", "nfs-reader",
		"-o", "jsonpath={.spec.volumes[*].nfs.server}")
	if err != nil {
		return fmt.Errorf("failed to get pod2: %w", err)
	}
	if strings.TrimSpace(output2) == "" {
		return fmt.Errorf("no NFS volume found on nfs-reader")
	}
	return nil
}

func (l *CKADVolumeNFSMountLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit nfs-writer pod", Command: "kubectl edit pod nfs-writer"},
		{Description: "Add NFS volume", Command: "Add nfs volume pointing to nfs-server.local:/exports/data"},
		{Description: "Edit nfs-reader pod", Command: "kubectl edit pod nfs-reader"},
		{Description: "Add same NFS volume", Command: "Add the same NFS volume configuration"},
		{Description: "Verify both pods have NFS", Command: "kubectl get pod nfs-writer nfs-reader -o yaml | grep nfs"},
	}
}
