package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&StatefulSetScalingLab{})
}

type StatefulSetScalingLab struct {
	BaseLab
}

func (l *StatefulSetScalingLab) ID() string { return "cka_statefulset_scaling" }
func (l *StatefulSetScalingLab) Title() string {
	return "Scale StatefulSet with PVC Management"
}
func (l *StatefulSetScalingLab) Category() Category     { return CategoryWorkloadsScheduling }
func (l *StatefulSetScalingLab) Difficulty() Difficulty { return DifficultyHard }
func (l *StatefulSetScalingLab) EstimatedTime() int     { return 25 }
func (l *StatefulSetScalingLab) Tags() []string {
	return []string{"statefulset", "scaling", "pvc", "storage"}
}
func (l *StatefulSetScalingLab) Cert() Cert        { return CertCKA }
func (l *StatefulSetScalingLab) DomainWeight() int { return 15 }

func (l *StatefulSetScalingLab) Description() string {
	return `A StatefulSet needs to be scaled down from 3 to 1 replicas, but PVCs
are not being cleaned up. Scale down the StatefulSet and ensure the
PersistentVolumeClaims are properly managed.`
}

func (l *StatefulSetScalingLab) Hints() []string {
	return []string{
		"Scale the StatefulSet using kubectl scale",
		"Check the PVC status after scaling",
		"Use volumeClaimTemplates for proper PVC management",
	}
}

func (l *StatefulSetScalingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *StatefulSetScalingLab) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: statefulset-ns
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
  namespace: statefulset-ns
spec:
  serviceName: nginx
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.27-alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, manifest); err != nil {
		return fmt.Errorf("creating statefulset: %w", err)
	}
	return nil
}

func (l *StatefulSetScalingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *StatefulSetScalingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "statefulset", "web",
		"-n", "statefulset-ns", "-o", "jsonpath={.spec.replicas}")
	if err != nil {
		return err
	}
	if output != "1" {
		return fmt.Errorf("StatefulSet replicas not scaled to 1")
	}
	return nil
}

func (l *StatefulSetScalingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current replicas", Command: "kubectl get statefulset web -n statefulset-ns"},
		{Description: "Scale down", Command: "kubectl scale statefulset web --replicas=1 -n statefulset-ns"},
		{Description: "Verify pods", Command: "kubectl get pods -n statefulset-ns"},
		{Description: "Check PVCs", Command: "kubectl get pvc -n statefulset-ns"},
	}
}
