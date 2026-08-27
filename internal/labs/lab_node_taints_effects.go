package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NodeTaintsEffectsLab{})
}

type NodeTaintsEffectsLab struct {
	BaseLab
}

func (l *NodeTaintsEffectsLab) ID() string {
	return "node_taints_effects"
}

func (l *NodeTaintsEffectsLab) Title() string {
	return "Node Taint with NoExecute Effect"
}

func (l *NodeTaintsEffectsLab) Category() Category {
	return CategoryScheduling
}

func (l *NodeTaintsEffectsLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *NodeTaintsEffectsLab) Description() string {
	return `A node has a taint with NoExecute effect: dedicated=reserved:NoExecute.
This not only prevents new pods from being scheduled but also evicts
existing pods without matching tolerations.

Your task: Add the correct toleration to a deployment so pods can
be scheduled on this node.`
}

func (l *NodeTaintsEffectsLab) Hints() []string {
	return []string{
		"Check node taints",
		"NoExecute effect evicts existing pods without tolerations",
		"The toleration must match key, value, and effect exactly",
	}
}

func (l *NodeTaintsEffectsLab) EstimatedTime() int {
	return 15
}

func (l *NodeTaintsEffectsLab) Tags() []string {
	return []string{"taints", "tolerations", "noexecute", "scheduling"}
}

func (l *NodeTaintsEffectsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeTaintsEffectsLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("finding node: %w", err)
	}

	if _, err := kubectl(ctx, kubeconfigPath, "taint", "node", nodeName,
		"dedicated=reserved:NoExecute"); err != nil {
		if !containsAny(err.Error(), "already has", "AlreadyExists") {
			return fmt.Errorf("tainting node: %w", err)
		}
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: reserved-app
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: reserved-app
  template:
    metadata:
      labels:
        app: reserved-app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo running; sleep 15; done']
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	return nil
}

func (l *NodeTaintsEffectsLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=reserved-app",
		"-o", "jsonpath={.items[*].status.phase}")
	for _, phase := range splitFields(output) {
		if phase == "Running" {
			return fmt.Errorf("expected pods to be pending")
		}
	}
	return nil
}

func (l *NodeTaintsEffectsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "reserved-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	if strings.TrimSpace(output) != "2" {
		return fmt.Errorf("deployment not ready (ready replicas: %s, expected: 2)", output)
	}

	return nil
}

func (l *NodeTaintsEffectsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check node taints",
			Command:     "kubectl describe nodes | grep Taints",
			Notes:       "Look for dedicated=reserved:NoExecute",
		},
		{
			Description: "Check pod status",
			Command:     "kubectl get pods -l app=reserved-app",
			Notes:       "Pods should be Pending due to taint",
		},
		{
			Description: "Fix: Add toleration",
			Command:     "kubectl patch deploy reserved-app --type='json' -p='[{\"op\":\"add\",\"path\":\"/spec/template/spec/tolerations\",\"value\":[{\"key\":\"dedicated\",\"operator\":\"Equal\",\"value\":\"reserved\",\"effect\":\"NoExecute\"}]}]'",
			Notes:       "Add toleration matching the taint exactly",
		},
		{
			Description: "Verify pods are running",
			Command:     "kubectl rollout status deploy/reserved-app",
			Notes:       "Both replicas should now be running",
		},
	}
}
