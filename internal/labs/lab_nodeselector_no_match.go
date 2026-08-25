package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&NodeSelectorNoMatchLab{}) }

type NodeSelectorNoMatchLab struct{ BaseLab }

func (l *NodeSelectorNoMatchLab) ID() string          { return "nodeselector_no_match" }
func (l *NodeSelectorNoMatchLab) Title() string        { return "Pod Pending — No Node Matches NodeSelector" }
func (l *NodeSelectorNoMatchLab) Category() Category   { return CategoryScheduling }
func (l *NodeSelectorNoMatchLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *NodeSelectorNoMatchLab) EstimatedTime() int   { return 10 }
func (l *NodeSelectorNoMatchLab) Tags() []string {
	return []string{"nodeselector", "scheduling", "pending"}
}
func (l *NodeSelectorNoMatchLab) Description() string {
	return `A pod 'specialized-app' is stuck Pending with event: "had taints that
the pod didn't tolerate" — but the real cause is a nodeSelector
requiring label disktype=ssd which no node in the cluster has.

Your task: Add the label disktype=ssd to any node so the pod can be
scheduled, OR fix the pod's nodeSelector. Choose the simpler path.`
}
func (l *NodeSelectorNoMatchLab) Hints() []string {
	return []string{
		"Check: kubectl describe pod specialized-app — look at Events",
		"Check: kubectl get nodes --show-labels — no node has disktype=ssd",
		"Fastest fix: kubectl label node <any-node> disktype=ssd",
		"The pod will be scheduled once at least one node matches",
	}
}

func (l *NodeSelectorNoMatchLab) Break(ctx context.Context, kp string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: specialized-app
  namespace: default
spec:
  nodeSelector:
    disktype: ssd
  containers:
  - name: app
    image: nginx:1.27-alpine
`
	return kubectlApply(ctx, kp, pod)
}

func (l *NodeSelectorNoMatchLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(3 * time.Second)
	return nil
}

func (l *NodeSelectorNoMatchLab) Verify(ctx context.Context, kp string) error {
	phase, _ := kubectl(ctx, kp, "get", "pod", "specialized-app", "-o", "jsonpath={.status.phase}")
	if phase != "Running" {
		return fmt.Errorf("pod not running (phase: %s)", phase)
	}
	nodes, _ := kubectl(ctx, kp, "get", "nodes", "-l", "disktype=ssd", "-o", "name")
	if nodes == "" {
		return fmt.Errorf("no node has disktype=ssd label")
	}
	return nil
}

func (l *NodeSelectorNoMatchLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod status", Command: "kubectl get pod specialized-app", Notes: "Shows Pending"},
		{Description: "Find cause", Command: "kubectl describe pod specialized-app | tail -10", Notes: "Events show no matching node for nodeSelector"},
		{Description: "Label a node", Command: "kubectl label node $(kubectl get nodes -o name | head -1) disktype=ssd", Notes: "Adds the required label to one node"},
		{Description: "Verify", Command: "kubectl get pod specialized-app", Notes: "Pod moves to Running"},
	}
}
