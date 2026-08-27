package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&PVDynamicProvisioningLab{})
}

type PVDynamicProvisioningLab struct {
	BaseLab
}

func (l *PVDynamicProvisioningLab) ID() string { return "cka_pv_dynamic_provisioning" }
func (l *PVDynamicProvisioningLab) Title() string {
	return "Configure Dynamic Volume Provisioning"
}
func (l *PVDynamicProvisioningLab) Category() Category     { return CategoryStorage }
func (l *PVDynamicProvisioningLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PVDynamicProvisioningLab) EstimatedTime() int     { return 20 }
func (l *PVDynamicProvisioningLab) Tags() []string {
	return []string{"dynamic-provisioning", "storageclass", "pvc"}
}
func (l *PVDynamicProvisioningLab) Cert() Cert        { return CertCKA }
func (l *PVDynamicProvisioningLab) DomainWeight() int { return 10 }

func (l *PVDynamicProvisioningLab) Description() string {
	return `Dynamic volume provisioning is not working because there is no default
StorageClass. Create a StorageClass with the local-path provisioner and
set it as the default.`
}

func (l *PVDynamicProvisioningLab) Hints() []string {
	return []string{
		"Create a StorageClass resource",
		"Set storageclass.kubernetes.io/is-default-class annotation",
		"Use a valid provisioner like k8s.io/no-provisioner or rancher.io/local-path",
	}
}

func (l *PVDynamicProvisioningLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PVDynamicProvisioningLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *PVDynamicProvisioningLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "storageclass", "-o",
		"jsonpath={.items[*].metadata.annotations.storageclass\\.kubernetes\\.io/is-default-class}")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "true") {
		return fmt.Errorf("no default StorageClass found")
	}
	return nil
}

func (l *PVDynamicProvisioningLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check StorageClasses", Command: "kubectl get storageclass"},
		{Description: "Create default StorageClass", Command: "cat <<EOF | kubectl apply -f -\napiVersion: storage.k8s.io/v1\nkind: StorageClass\nmetadata:\n  name: default-storage\n  annotations:\n    storageclass.kubernetes.io/is-default-class: \"true\"\nprovisioner: rancher.io/local-path\nreclaimPolicy: Delete\nvolumeBindingMode: WaitForFirstConsumer\nEOF"},
		{Description: "Verify", Command: "kubectl get storageclass"},
	}
}
