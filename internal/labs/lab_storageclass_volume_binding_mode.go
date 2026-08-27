package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&StorageClassVolumeBindingModeLab{})
}

type StorageClassVolumeBindingModeLab struct {
	BaseLab
}

func (l *StorageClassVolumeBindingModeLab) ID() string {
	return "storageclass_volume_binding_mode"
}

func (l *StorageClassVolumeBindingModeLab) Title() string {
	return "Volume Binding Mode Issue"
}

func (l *StorageClassVolumeBindingModeLab) Category() Category {
	return CategoryStorage
}

func (l *StorageClassVolumeBindingModeLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *StorageClassVolumeBindingModeLab) Description() string {
	return `A StorageClass 'local-storage' has WaitForFirstConsumer binding mode
but the PersistentVolumeClaim is not being bound because no consumer
(Pod) exists yet. This creates a chicken-and-egg problem.

Your task: Create a Pod that uses the PVC to trigger volume binding.`
}

func (l *StorageClassVolumeBindingModeLab) Hints() []string {
	return []string{
		"Check the StorageClass binding mode",
		"WaitForFirstConsumer delays binding until a Pod uses the PVC",
		"Create a Pod that references the PVC",
	}
}

func (l *StorageClassVolumeBindingModeLab) EstimatedTime() int {
	return 15
}

func (l *StorageClassVolumeBindingModeLab) Tags() []string {
	return []string{"storageclass", "volume-binding", "pvc", "storage"}
}

func (l *StorageClassVolumeBindingModeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *StorageClassVolumeBindingModeLab) Break(ctx context.Context, kubeconfigPath string) error {
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
      storage: 5Gi
  storageClassName: local-storage
`
	if err := kubectlApply(ctx, kubeconfigPath, pvc); err != nil {
		return fmt.Errorf("creating PVC: %w", err)
	}

	return nil
}

func (l *StorageClassVolumeBindingModeLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pvc", "local-pvc",
		"-o", "jsonpath={.status.phase}")
	if strings.TrimSpace(output) == "Pending" {
		return nil
	}
	return nil
}

func (l *StorageClassVolumeBindingModeLab) Verify(ctx context.Context, kubeconfigPath string) error {
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

func (l *StorageClassVolumeBindingModeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check StorageClass",
			Command:     "kubectl get sc local-storage -o yaml | grep volumeBindingMode",
			Notes:       "bindingMode is WaitForFirstConsumer",
		},
		{
			Description: "Check PVC status",
			Command:     "kubectl get pvc local-pvc",
			Notes:       "PVC is Pending, waiting for a consumer",
		},
		{
			Description: "Create a Pod using the PVC",
			Command:     "cat <<EOF | kubectl apply -f -\napiVersion: v1\nkind: Pod\nmetadata:\n  name: local-pod\n  namespace: default\nspec:\n  containers:\n  - name: app\n    image: busybox:1.36\n    command: ['sh', '-c', 'sleep 3600']\n    volumeMounts:\n    - name: data\n      mountPath: /data\n  volumes:\n  - name: data\n    persistentVolumeClaim:\n      claimName: local-pvc\nEOF",
			Notes:       "Pod triggers the PVC binding",
		},
		{
			Description: "Verify PVC is Bound",
			Command:     "kubectl get pvc local-pvc",
			Notes:       "Should now show Bound status",
		},
	}
}
