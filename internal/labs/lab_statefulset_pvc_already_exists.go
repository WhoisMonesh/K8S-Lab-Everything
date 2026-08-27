package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&StatefulSetPVCAlreadyExistsLab{})
}

type StatefulSetPVCAlreadyExistsLab struct {
	BaseLab
}

func (l *StatefulSetPVCAlreadyExistsLab) ID() string {
	return "statefulset_pvc_already_exists"
}

func (l *StatefulSetPVCAlreadyExistsLab) Title() string {
	return "StatefulSet PVC Naming Conflict"
}

func (l *StatefulSetPVCAlreadyExistsLab) Category() Category {
	return CategoryWorkloads
}

func (l *StatefulSetPVCAlreadyExistsLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *StatefulSetPVCAlreadyExistsLab) Description() string {
	return `A StatefulSet 'redis-cluster' cannot create pods because PVCs with
incorrect names already exist. The StatefulSet expects PVCs named
redis-cluster-0, redis-cluster-1, etc. but someone manually created
PVCs with different names that conflict.

Your task: Fix the PVC naming conflict so the StatefulSet can create pods.`
}

func (l *StatefulSetPVCAlreadyExistsLab) Hints() []string {
	return []string{
		"Check the StatefulSet pod status",
		"List PVCs in the namespace",
		"StatefulSet PVCs follow the pattern <statefulset-name>-<ordinal>",
		"You may need to delete incorrectly named PVCs",
	}
}

func (l *StatefulSetPVCAlreadyExistsLab) EstimatedTime() int {
	return 20
}

func (l *StatefulSetPVCAlreadyExistsLab) Tags() []string {
	return []string{"statefulset", "pvc", "naming", "workloads"}
}

func (l *StatefulSetPVCAlreadyExistsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *StatefulSetPVCAlreadyExistsLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create PVCs with wrong names first
	wrongPVCs := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-0
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: standard
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-1
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
  storageClassName: standard
`
	if err := kubectlApply(ctx, kubeconfigPath, wrongPVCs); err != nil {
		return fmt.Errorf("creating wrong PVCs: %w", err)
	}

	// Create StatefulSet expecting redis-cluster-<ordinal> PVCs
	statefulset := `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-cluster
  namespace: default
spec:
  serviceName: redis-headless
  replicas: 2
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7
        ports:
        - containerPort: 6379
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
      storageClassName: standard
`
	if err := kubectlApply(ctx, kubeconfigPath, statefulset); err != nil {
		return fmt.Errorf("creating statefulset: %w", err)
	}

	// Create headless service
	service := `apiVersion: v1
kind: Service
metadata:
  name: redis-headless
  namespace: default
spec:
  clusterIP: None
  selector:
    app: redis
  ports:
  - port: 6379
`
	if err := kubectlApply(ctx, kubeconfigPath, service); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}

func (l *StatefulSetPVCAlreadyExistsLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *StatefulSetPVCAlreadyExistsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "statefulset", "redis-cluster",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check statefulset: %w", err)
	}

	if strings.TrimSpace(output) != "2" {
		return fmt.Errorf("statefulset not ready (ready replicas: %s, expected: 2)", output)
	}

	return nil
}

func (l *StatefulSetPVCAlreadyExistsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check StatefulSet status",
			Command:     "kubectl get statefulset redis-cluster",
			Notes:       "READY column should show 0/2",
		},
		{
			Description: "Check StatefulSet events",
			Command:     "kubectl describe statefulset redis-cluster | grep -A 10 Events",
			Notes:       "Look for PVC already exists errors",
		},
		{
			Description: "List PVCs",
			Command:     "kubectl get pvc",
			Notes:       "You'll see redis-0 and redis-1 instead of redis-cluster-data-0 and redis-cluster-data-1",
		},
		{
			Description: "Delete the incorrectly named PVCs",
			Command:     "kubectl delete pvc redis-0 redis-1",
			Notes:       "Remove the PVCs with wrong names",
		},
		{
			Description: "Wait for StatefulSet to create correct PVCs",
			Command:     "kubectl rollout status statefulset/redis-cluster --timeout=120s",
			Notes:       "StatefulSet will create redis-cluster-data-0 and redis-cluster-data-1",
		},
	}
}
