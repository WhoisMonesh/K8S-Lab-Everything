package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&NodeTaintNoTolerationLab{})
}

type NodeTaintNoTolerationLab struct {
	BaseLab
}

func (l *NodeTaintNoTolerationLab) ID() string {
	return "taint_no_toleration"
}

func (l *NodeTaintNoTolerationLab) Title() string {
	return "Pods Repelled By Node Taint"
}

func (l *NodeTaintNoTolerationLab) Category() Category {
	return CategoryScheduling
}

func (l *NodeTaintNoTolerationLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *NodeTaintNoTolerationLab) Description() string {
	return `The cluster administrator tainted the worker node with
dedicated=processing:NoSchedule to reserve it, but forgot to update the
workload. A deployment named 'batch-worker' now cannot place any of its
2 replicas - they all sit in Pending.

Your task: Make the batch-worker pods schedulable again by adding proper
tolerations (removing the taint is also acceptable here).`
}

func (l *NodeTaintNoTolerationLab) Hints() []string {
	return []string{
		"kubectl describe pod shows '0/N nodes are available' and names the taint",
		"Inspect node taints with kubectl get nodes -o custom-columns or kubectl describe node",
		"A toleration must match key, value and effect exactly to admit the pod",
		"The tolerations field lives at spec.template.spec.tolerations on the deployment",
	}
}

func (l *NodeTaintNoTolerationLab) EstimatedTime() int {
	return 20
}

func (l *NodeTaintNoTolerationLab) Tags() []string {
	return []string{"taints", "tolerations", "scheduling", "pending", "troubleshooting"}
}

func (l *NodeTaintNoTolerationLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeTaintNoTolerationLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("finding node to taint: %w", err)
	}

	if _, err := kubectl(ctx, kubeconfigPath, "taint", "node", nodeName,
		"dedicated=processing:NoSchedule"); err != nil {
		// Tolerate re-runs where the taint already exists
		if !containsAny(err.Error(), "already has", "AlreadyExists") {
			return fmt.Errorf("tainting node: %w", err)
		}
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: batch-worker
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: batch-worker
  template:
    metadata:
      labels:
        app: batch-worker
    spec:
      containers:
      - name: worker
        image: busybox:1.36
        command: ["sh", "-c", "while true; do echo crunching; sleep 15; done"]
`

	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating pending deployment: %w", err)
	}
	return nil
}

func (l *NodeTaintNoTolerationLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=batch-worker",
		"-o", "jsonpath={.items[*].status.phase}")
	for _, phase := range splitFields(output) {
		if phase == "Running" {
			return fmt.Errorf("expected pods to be Pending")
		}
	}
	return nil
}

func (l *NodeTaintNoTolerationLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "batch-worker",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}
	if output != "2" {
		return fmt.Errorf("deployment not fully ready yet (ready replicas: %s, expected: 2)", output)
	}
	return nil
}

func (l *NodeTaintNoTolerationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "See why scheduling fails",
			Command:     "kubectl describe pod -l app=batch-worker | grep -A 5 Events",
			Notes:       "'node(s) had untolerated taint {dedicated: processing}' confirms the cause",
		},
		{
			Description: "List the taints on the nodes",
			Command:     "kubectl get nodes -o jsonpath='{range .items[*]}{.metadata.name}{\"\\t\"}{.spec.taints}{\"\\n\"}{end}'",
			Notes:       "You will see dedicated=processing:NoSchedule on the node",
		},
		{
			Description: "Fix option A: add a matching toleration",
			Command:     "kubectl patch deploy batch-worker --type='json' -p='[{\"op\":\"add\",\"path\":\"/spec/template/spec/tolerations\",\"value\":[{\"key\":\"dedicated\",\"operator\":\"Equal\",\"value\":\"processing\",\"effect\":\"NoSchedule\"}]}]'",
			Notes:       "Key, value and effect must all match the taint",
		},
		{
			Description: "Fix option B: remove the taint from the node",
			Command:     "kubectl taint node <node-name> dedicated=processing:NoSchedule-",
			Notes:       "Trailing minus removes the taint; fine in this lab but loses the reservation in real clusters",
		},
		{
			Description: "Confirm both replicas run",
			Command:     "kubectl rollout status deploy/batch-worker && kubectl get pods -l app=batch-worker",
			Notes:       "Both pods should reach Running",
		},
	}
}
