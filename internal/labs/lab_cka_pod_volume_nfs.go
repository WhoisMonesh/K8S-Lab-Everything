package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodVolumeNFSLab{})
}

type PodVolumeNFSLab struct {
	BaseLab
}

func (l *PodVolumeNFSLab) ID() string             { return "cka_pod_volume_nfs" }
func (l *PodVolumeNFSLab) Title() string          { return "Mount NFS Volume to Pod" }
func (l *PodVolumeNFSLab) Category() Category     { return CategoryStorage }
func (l *PodVolumeNFSLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PodVolumeNFSLab) EstimatedTime() int     { return 20 }
func (l *PodVolumeNFSLab) Tags() []string {
	return []string{"nfs", "volume", "storage", "pod"}
}
func (l *PodVolumeNFSLab) Cert() Cert        { return CertCKA }
func (l *PodVolumeNFSLab) DomainWeight() int { return 10 }

func (l *PodVolumeNFSLab) Description() string {
	return `A pod needs to mount an NFS share for shared data. Configure the pod
to use an NFS volume pointing to nfs-server:/exports with read-write access.`
}

func (l *PodVolumeNFSLab) Hints() []string {
	return []string{
		"Add an nfs volume to the pod spec",
		"Mount the volume in the container",
		"Ensure the NFS server address and path are correct",
	}
}

func (l *PodVolumeNFSLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodVolumeNFSLab) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: nfs-ns
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs-pvc
  namespace: nfs-ns
spec:
  accessModes:
  - ReadWriteOnce
  storageClassName: nfs-storage
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: Pod
metadata:
  name: nfs-pod
  namespace: nfs-ns
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'while true; do echo nfs; sleep 15; done']
    volumeMounts:
    - name: nfs-volume
      mountPath: /data
  volumes:
  - name: nfs-volume
    nfs:
      server: wrong-nfs-server
      path: /wrong
`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PodVolumeNFSLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodVolumeNFSLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "nfs-pod",
		"-n", "nfs-ns", "-o", "jsonpath={.spec.volumes}")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "nfs") {
		return fmt.Errorf("NFS volume not configured")
	}
	return nil
}

func (l *PodVolumeNFSLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod spec", Command: "kubectl get pod nfs-pod -n nfs-ns -o yaml"},
		{Description: "Add NFS volume", Command: "Add to volumes:\n- name: nfs-volume\n  nfs:\n    server: nfs-server\n    path: /exports"},
		{Description: "Mount volume", Command: "Add volumeMounts:\n- name: nfs-volume\n  mountPath: /data"},
		{Description: "Verify", Command: "kubectl get pod nfs-pod -n nfs-ns -o yaml"},
	}
}
