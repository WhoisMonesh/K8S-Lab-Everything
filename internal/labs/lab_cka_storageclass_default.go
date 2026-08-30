package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&StorageClassDefaultLab{})
}

type StorageClassDefaultLab struct {
	BaseLab
}

func (l *StorageClassDefaultLab) ID() string { return "cka_storageclass_default" }
func (l *StorageClassDefaultLab) Title() string {
	return "Set Default StorageClass"
}
func (l *StorageClassDefaultLab) Category() Category     { return CategoryStorage }
func (l *StorageClassDefaultLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *StorageClassDefaultLab) EstimatedTime() int     { return 15 }
func (l *StorageClassDefaultLab) Tags() []string {
	return []string{"storageclass", "default", "storage"}
}
func (l *StorageClassDefaultLab) Cert() Cert        { return CertCKA }
func (l *StorageClassDefaultLab) DomainWeight() int { return 10 }

func (l *StorageClassDefaultLab) Description() string {
	return `There are multiple StorageClasses but none is marked as default. Set
the "standard" StorageClass as the default for dynamic provisioning.`
}

func (l *StorageClassDefaultLab) Hints() []string {
	return []string{
		"Annotate the StorageClass with is-default-class: true",
		"Remove default annotation from other StorageClasses",
		"Use kubectl annotate to update",
	}
}

func (l *StorageClassDefaultLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *StorageClassDefaultLab) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
provisioner: rancher.io/local-path
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast
provisioner: rancher.io/local-path
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *StorageClassDefaultLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *StorageClassDefaultLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "storageclass", "standard",
		"-o", "jsonpath={.metadata.annotations.storageclass\\.kubernetes\\.io/is-default-class}")
	if err != nil {
		return err
	}
	if output != "true" {
		return fmt.Errorf("standard StorageClass is not default")
	}
	return nil
}

func (l *StorageClassDefaultLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check StorageClasses", Command: "kubectl get storageclass"},
		{Description: "Annotate as default", Command: "kubectl annotate storageclass standard storageclass.kubernetes.io/is-default-class=true --overwrite"},
		{Description: "Verify", Command: "kubectl get storageclass"},
	}
}
