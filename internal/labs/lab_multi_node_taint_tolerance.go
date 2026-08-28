package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&MultiNodeTaintToleranceLab{})
}

type MultiNodeTaintToleranceLab struct {
	BaseLab
}

func (l *MultiNodeTaintToleranceLab) ID() string {
	return "multi_node_taint_tolerance"
}

func (l *MultiNodeTaintToleranceLab) Title() string {
	return "Multi-Node Taint and Toleration"
}

func (l *MultiNodeTaintToleranceLab) Category() Category {
	return CategoryWorkloadsScheduling
}

func (l *MultiNodeTaintToleranceLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *MultiNodeTaintToleranceLab) Description() string {
	return `Worker nodes have been tainted for dedicated workloads. A deployment
needs to run on these tainted nodes but lacks the proper tolerations.

Your task: Add the correct tolerations to the deployment so pods can be
scheduled on the tainted worker nodes.`
}

func (l *MultiNodeTaintToleranceLab) Hints() []string {
	return []string{
		"Check node taints across all nodes",
		"The toleration key, value, and effect must match exactly",
		"Use kubectl describe node to see taint details",
	}
}

func (l *MultiNodeTaintToleranceLab) EstimatedTime() int {
	return 15
}

func (l *MultiNodeTaintToleranceLab) Tags() []string {
	return []string{"taints", "tolerations", "multi-node", "scheduling", "dedicated-nodes"}
}

func (l *MultiNodeTaintToleranceLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	nodes, err := kubectl(ctx, kubeconfigPath, "get", "nodes",
		"-l", "!node-role.kubernetes.io/control-plane",
		"-o", "jsonpath={.items[0].metadata.name}")
	if err != nil {
		return fmt.Errorf("getting worker node: %w", err)
	}

	workerNode := strings.TrimSpace(nodes)
	if workerNode == "" {
		return fmt.Errorf("no worker nodes found")
	}

	if _, err := kubectl(ctx, kubeconfigPath, "taint", "node", workerNode,
		"dedicated=gpu:NoSchedule"); err != nil {
		return fmt.Errorf("tainting node: %w", err)
	}

	return nil
}

func (l *MultiNodeTaintToleranceLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: gpu-workload
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: gpu-workload
  template:
    metadata:
      labels:
        app: gpu-workload
    spec:
      nodeSelector:
        node-role.kubernetes.io/worker: ""
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo gpu; sleep 15; done']
        resources:
          limits:
            memory: 32Mi
            cpu: 50m
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *MultiNodeTaintToleranceLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=gpu-workload",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("checking pods: %w", err)
	}

	if strings.Contains(output, "Running") {
		return fmt.Errorf("pods are running (expected pending due to taint)")
	}

	return nil
}

func (l *MultiNodeTaintToleranceLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "gpu-workload",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("checking deployment: %w", err)
	}

	ready := strings.TrimSpace(output)
	if ready != "2" {
		return fmt.Errorf("deployment not ready (ready: %s, expected: 2)", ready)
	}

	return nil
}

func (l *MultiNodeTaintToleranceLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check node taints",
			Command:     "kubectl describe nodes | grep -A 5 Taints",
			Notes:       "Find dedicated=gpu:NoSchedule",
		},
		{
			Description: "Fix: Add toleration to deployment",
			Command:     `kubectl patch deploy gpu-workload --type='json' -p='[{"op":"add","path":"/spec/template/spec/tolerations","value":[{"key":"dedicated","operator":"Equal","value":"gpu","effect":"NoSchedule"}]}]'`,
			Notes:       "Match the taint key, value, and effect",
		},
		{
			Description: "Verify pods are running",
			Command:     "kubectl rollout status deploy/gpu-workload",
			Notes:       "Both replicas should be ready on the tainted node",
		},
	}
}
