package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PVLocalNodeAffinityLab{})
}

type PVLocalNodeAffinityLab struct {
	BaseLab
}

func (l *PVLocalNodeAffinityLab) ID() string {
	return "pv_local_node_affinity"
}

func (l *PVLocalNodeAffinityLab) Title() string {
	return "Local PV Node Affinity Mismatch"
}

func (l *PVLocalNodeAffinityLab) Category() Category {
	return CategoryStorage
}

func (l *PVLocalNodeAffinityLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *PVLocalNodeAffinityLab) Description() string {
	return `A Local PersistentVolume 'local-pv' has node affinity set to
node-name=worker-2 but the pod requesting it is scheduled on worker-1.
The PVC remains Pending because the node affinity doesn't match.

Your task: Fix the PV node affinity or ensure the pod is scheduled
on the correct node.`
}

func (l *PVLocalNodeAffinityLab) Hints() []string {
	return []string{
		"Check the PV node affinity",
		"Local PVs are only available on specific nodes",
		"Either change PV affinity or add nodeSelector to the pod",
	}
}

func (l *PVLocalNodeAffinityLab) EstimatedTime() int {
	return 20
}

func (l *PVLocalNodeAffinityLab) Tags() []string {
	return []string{"local-pv", "node-affinity", "storage"}
}

func (l *PVLocalNodeAffinityLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PVLocalNodeAffinityLab) Break(ctx context.Context, kubeconfigPath string) error {
	pv := `apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-pv
spec:
  capacity:
    storage: 10Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /mnt/data
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - worker-2
`
	if err := kubectlApply(ctx, kubeconfigPath, pv); err != nil {
		return fmt.Errorf("creating PV: %w", err)
	}

	storageClass := `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Delete
`
	if err := kubectlApply(ctx, kubeconfigPath, storageClass); err != nil {
		return fmt.Errorf("creating StorageClass: %w", err)
	}

	pvc := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: local-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: local-storage
`
	if err := kubectlApply(ctx, kubeconfigPath, pvc); err != nil {
		return fmt.Errorf("creating PVC: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: local-pod
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'sleep 3600']
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: local-pvc
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PVLocalNodeAffinityLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pvc", "local-pvc",
		"-o", "jsonpath={.status.phase}")
	if strings.TrimSpace(output) == "Pending" {
		return nil
	}
	return nil
}

func (l *PVLocalNodeAffinityLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pvc", "local-pvc",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check PVC: %w", err)
	}

	if strings.TrimSpace(output) != "Bound" {
		return fmt.Errorf("PVC is not Bound (status: %s)", output)
	}

	return nil
}

func (l *PVLocalNodeAffinityLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check PV node affinity",
			Command:     "kubectl get pv local-pv -o yaml | grep -A 10 nodeAffinity",
			Notes:       "PV requires worker-2 but pod might be on worker-1",
		},
		{
			Description: "Check PVC status",
			Command:     "kubectl get pvc local-pvc",
			Notes:       "PVC is Pending, waiting for matching PV",
		},
		{
			Description: "Fix: Add nodeSelector to pod",
			Command:     "kubectl edit pod local-pod",
			Notes:       "Add spec.nodeSelector with kubernetes.io/hostname: worker-2",
		},
		{
			Description: "Alternative: Fix PV affinity",
			Command:     "kubectl edit pv local-pv",
			Notes:       "Change node affinity to match the actual node",
		},
		{
			Description: "Verify PVC is Bound",
			Command:     "kubectl get pvc local-pvc",
			Notes:       "Should now show Bound status",
		},
	}
}
