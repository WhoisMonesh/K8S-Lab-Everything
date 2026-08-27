package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&NodeCordonedLab{}) }

type NodeCordonedLab struct{ BaseLab }

func (l *NodeCordonedLab) ID() string             { return "node_cordoned" }
func (l *NodeCordonedLab) Title() string          { return "Node Cordoned — Pods Cannot Schedule" }
func (l *NodeCordonedLab) Category() Category     { return CategoryControlPlane }
func (l *NodeCordonedLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *NodeCordonedLab) EstimatedTime() int     { return 10 }
func (l *NodeCordonedLab) Tags() []string {
	return []string{"node", "cordon", "scheduling", "control-plane"}
}
func (l *NodeCordonedLab) Description() string {
	return `The only worker node in the cluster has been cordoned (marked
unschedulable). A new deployment 'web' cannot schedule any pods.

Your task: Uncordon the node so the deployment can schedule.`
}
func (l *NodeCordonedLab) Hints() []string {
	return []string{
		"Check: kubectl get nodes — look for SchedulingDisabled status",
		"Check: kubectl describe node <name> | grep Taints",
		"Uncordon: kubectl uncordon <node-name>",
	}
}

func (l *NodeCordonedLab) Break(ctx context.Context, kp string) error {
	node, err := kubectl(ctx, kp, "get", "nodes", "-l", "!node-role.kubernetes.io/control-plane",
		"-o", "jsonpath={.items[0].metadata.name}")
	if err != nil || strings.TrimSpace(node) == "" {
		node, err = kubectl(ctx, kp, "get", "nodes", "-o", "jsonpath={.items[0].metadata.name}")
		if err != nil {
			return fmt.Errorf("cannot find a node to cordon")
		}
	}
	node = strings.TrimSpace(node)
	if _, err := kubectl(ctx, kp, "cordon", node); err != nil {
		return err
	}
	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: nginx
        image: nginx:1.27-alpine
`
	return kubectlApply(ctx, kp, deploy)
}

func (l *NodeCordonedLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *NodeCordonedLab) Verify(ctx context.Context, kp string) error {
	out, _ := kubectl(ctx, kp, "get", "nodes", "-o", "jsonpath={.items[*].spec.unschedulable}")
	if strings.Contains(out, "true") {
		return fmt.Errorf("node is still cordoned (unschedulable=true)")
	}
	ready, _ := kubectl(ctx, kp, "get", "deploy", "web", "-o", "jsonpath={.status.readyReplicas}")
	if ready != "2" {
		return fmt.Errorf("deployment not ready (ready: %s)", ready)
	}
	return nil
}

func (l *NodeCordonedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check node status", Command: "kubectl get nodes", Notes: "One node shows SchedulingDisabled"},
		{Description: "Check the taint", Command: "kubectl describe node $(kubectl get nodes -o name | head -1) | grep -A5 Taints", Notes: "Has node.kubernetes.io/unschedulable:NoSchedule"},
		{Description: "Uncordon", Command: "kubectl uncordon $(kubectl get nodes -o name | head -1)", Notes: "Removes the unschedulable taint"},
		{Description: "Verify pods schedule", Command: "kubectl get pods -l app=web", Notes: "Pods move from Pending to Running"},
	}
}
