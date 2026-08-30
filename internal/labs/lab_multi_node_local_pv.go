package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&MultiNodeLocalPVLab{})
}

type MultiNodeLocalPVLab struct {
	BaseLab
}

func (l *MultiNodeLocalPVLab) ID() string {
	return "multi_node_local_pv"
}

func (l *MultiNodeLocalPVLab) Title() string {
	return "Local PersistentVolume Node Affinity"
}

func (l *MultiNodeLocalPVLab) Category() Category {
	return CategoryStorage
}

func (l *MultiNodeLocalPVLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *MultiNodeLocalPVLab) Description() string {
	return `A Local PersistentVolume is configured but pods using it are stuck
in Pending state. The PV has node affinity pointing to a non-existent node.

Your task: Fix the PersistentVolume's node affinity to point to an actual
worker node so the pod can be scheduled.`
}

func (l *MultiNodeLocalPVLab) Hints() []string {
	return []string{
		"Check the PV nodeAffinity configuration",
		"The node name in the affinity doesn't match any real node",
		"Get the actual worker node name and update the PV",
	}
}

func (l *MultiNodeLocalPVLab) EstimatedTime() int {
	return 20
}

func (l *MultiNodeLocalPVLab) Tags() []string {
	return []string{"persistent-volume", "local-pv", "node-affinity", "storage", "multi-node"}
}

func (l *MultiNodeLocalPVLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

// ClusterSpec declares a multi-worker cluster so node scheduling/scaling
// scenarios are real on kind.
func (l *MultiNodeLocalPVLab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           2,
	}
}

func (l *MultiNodeLocalPVLab) Break(ctx context.Context, kubeconfigPath string) error {
	pv := `apiVersion: v1
kind: PersistentVolume
metadata:
  name: local-pv
spec:
  capacity:
    storage: 1Gi
  volumeMode: Filesystem
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-storage
  local:
    path: /tmp/data
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - worker-does-not-exist
`
	if err := kubectlApply(ctx, kubeconfigPath, pv); err != nil {
		return err
	}

	pvc := `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: local-pvc
spec:
  accessModes:
  - ReadWriteOnce
  storageClassName: local-storage
  resources:
    requests:
      storage: 1Gi
`
	if err := kubectlApply(ctx, kubeconfigPath, pvc); err != nil {
		return err
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: local-pod
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'while true; do echo local; sleep 15; done']
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: local-pvc
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *MultiNodeLocalPVLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pv", "local-pv",
		"-o", "jsonpath={.spec.nodeAffinity.required.nodeSelectorTerms[0].matchExpressions[0].values[0]}")
	if err != nil {
		return fmt.Errorf("checking pv: %w", err)
	}

	if strings.TrimSpace(output) == "worker-does-not-exist" {
		return nil
	}

	return fmt.Errorf("PV node affinity points to existing node (expected broken)")
}

func (l *MultiNodeLocalPVLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pv", "local-pv",
		"-o", "jsonpath={.spec.nodeAffinity.required.nodeSelectorTerms[0].matchExpressions[0].values[0]}")
	if err != nil {
		return fmt.Errorf("checking pv: %w", err)
	}

	if strings.TrimSpace(output) == "worker-does-not-exist" {
		return fmt.Errorf("PV still points to non-existent node")
	}

	return nil
}

func (l *MultiNodeLocalPVLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check PV node affinity",
			Command:     "kubectl get pv local-pv -o yaml | grep -A 10 nodeAffinity",
			Notes:       "Points to worker-does-not-exist",
		},
		{
			Description: "Get actual worker node name",
			Command:     "kubectl get nodes -l '!node-role.kubernetes.io/control-plane' -o jsonpath='{.items[0].metadata.name}'",
			Notes:       "Use the real worker node name",
		},
		{
			Description: "Fix: Update PV node affinity",
			Command:     "kubectl patch pv local-pv --type json -p '[{\"op\":\"replace\",\"path\":\"/spec/nodeAffinity/required/nodeSelectorTerms/0/matchExpressions/0/values/0\",\"value\":\"<real-worker-node>\"}]'",
			Notes:       "Replace <real-worker-node> with actual node name",
		},
		{
			Description: "Verify pod is scheduled",
			Command:     "kubectl get pod local-pod -o wide",
			Notes:       "Pod should be running on the correct node",
		},
	}
}
