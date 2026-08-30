package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&StatefulSetVolumeClaimsLab{})
}

type StatefulSetVolumeClaimsLab struct {
	BaseLab
}

func (l *StatefulSetVolumeClaimsLab) ID() string {
	return "cka_statefulset_volume_claims"
}
func (l *StatefulSetVolumeClaimsLab) Title() string {
	return "Use volumeClaimTemplates"
}
func (l *StatefulSetVolumeClaimsLab) Category() Category     { return CategoryStorage }
func (l *StatefulSetVolumeClaimsLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *StatefulSetVolumeClaimsLab) EstimatedTime() int     { return 20 }
func (l *StatefulSetVolumeClaimsLab) Tags() []string {
	return []string{"statefulset", "volumeclaimtemplates", "pvc", "storage"}
}
func (l *StatefulSetVolumeClaimsLab) Cert() Cert        { return CertCKA }
func (l *StatefulSetVolumeClaimsLab) DomainWeight() int { return 10 }

func (l *StatefulSetVolumeClaimsLab) Description() string {
	return `A StatefulSet is running without persistent storage. Add a
volumeClaimTemplate to the StatefulSet so each replica gets its own
PersistentVolumeClaim for data persistence.`
}

func (l *StatefulSetVolumeClaimsLab) Hints() []string {
	return []string{
		"Add volumeClaimTemplates to StatefulSet spec",
		"Define PVC spec with accessModes and resources",
		"Mount the volume in the container spec",
	}
}

func (l *StatefulSetVolumeClaimsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *StatefulSetVolumeClaimsLab) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: ss-ns
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: data-app
  namespace: ss-ns
spec:
  serviceName: data-app
  replicas: 1
  selector:
    matchLabels:
      app: data-app
  template:
    metadata:
      labels:
        app: data-app
    spec:
      containers:
      - name: app
        image: nginx:1.27-alpine
`
	if err := kubectlApply(ctx, kubeconfigPath, manifest); err != nil {
		return fmt.Errorf("creating statefulset without storage: %w", err)
	}
	return nil
}

func (l *StatefulSetVolumeClaimsLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *StatefulSetVolumeClaimsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "statefulset", "data-app",
		"-n", "ss-ns", "-o", "jsonpath={.spec.volumeClaimTemplates}")
	if err != nil {
		return err
	}
	if output == "" || output == "null" {
		return fmt.Errorf("volumeClaimTemplates not configured")
	}
	return nil
}

func (l *StatefulSetVolumeClaimsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check StatefulSet", Command: "kubectl get statefulset data-app -n ss-ns -o yaml"},
		{Description: "Patch to add volumeClaimTemplates", Command: "cat <<EOF | kubectl patch statefulset data-app -n ss-ns --type=strategic -p -\nspec:\n  volumeClaimTemplates:\n  - metadata:\n      name: data\n    spec:\n      accessModes: [\"ReadWriteOnce\"]\n      resources:\n        requests:\n          storage: 5Gi\nEOF"},
		{Description: "Verify PVCs", Command: "kubectl get pvc -n ss-ns"},
	}
}
