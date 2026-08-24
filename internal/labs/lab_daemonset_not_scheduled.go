package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&DaemonSetNotScheduledLab{})
}

type DaemonSetNotScheduledLab struct {
	BaseLab
}

func (l *DaemonSetNotScheduledLab) ID() string {
	return "daemonset_not_scheduled"
}

func (l *DaemonSetNotScheduledLab) Title() string {
	return "DaemonSet Not Scheduling on Nodes"
}

func (l *DaemonSetNotScheduledLab) Category() Category {
	return CategoryWorkloads
}

func (l *DaemonSetNotScheduledLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *DaemonSetNotScheduledLab) Description() string {
	return `A DaemonSet has been deployed to run a monitoring agent on all nodes,
but the pods are not being scheduled. The DaemonSet shows 0 pods ready
even though nodes are available.

Your task: Diagnose why the DaemonSet pods are not being scheduled and
fix the issue so that the monitoring agent runs on all nodes.`
}

func (l *DaemonSetNotScheduledLab) Hints() []string {
	return []string{
		"Check the DaemonSet configuration for node selectors or affinity rules",
		"Look at node labels to see if they match the DaemonSet requirements",
		"Check for taints on the nodes that might prevent scheduling",
		"Verify the tolerations in the DaemonSet spec",
	}
}

func (l *DaemonSetNotScheduledLab) EstimatedTime() int {
	return 20
}

func (l *DaemonSetNotScheduledLab) Tags() []string {
	return []string{"daemonset", "scheduling", "node-selector", "workloads", "troubleshooting"}
}

func (l *DaemonSetNotScheduledLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	// Create a namespace
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: monitoring
`
	return kubectlApply(ctx, kubeconfigPath, namespace)
}

func (l *DaemonSetNotScheduledLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create a DaemonSet with a node selector that doesn't match any nodes
	brokenDaemonSet := `apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: monitoring-agent
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: monitoring-agent
  template:
    metadata:
      labels:
        app: monitoring-agent
    spec:
      nodeSelector:
        node-type: monitoring
      containers:
      - name: agent
        image: busybox:latest
        command: ["sh", "-c", "while true; do echo monitoring...; sleep 30; done"]
        resources:
          limits:
            memory: 50Mi
            cpu: 50m
`
	return kubectlApply(ctx, kubeconfigPath, brokenDaemonSet)
}

func (l *DaemonSetNotScheduledLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Give it a moment
	time.Sleep(5 * time.Second)

	// Check if DaemonSet has no pods scheduled
	output, err := kubectl(ctx, kubeconfigPath, "get", "daemonset",
		"monitoring-agent", "-n", "monitoring",
		"-o", "jsonpath={.status.numberReady}")
	if err != nil {
		return fmt.Errorf("DaemonSet not found: %w", err)
	}

	if strings.TrimSpace(output) != "0" && strings.TrimSpace(output) != "" {
		return fmt.Errorf("DaemonSet has pods scheduled (should be 0)")
	}

	return nil
}

func (l *DaemonSetNotScheduledLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Wait a bit for pods to potentially start
	time.Sleep(10 * time.Second)

	// Check if DaemonSet now has pods scheduled
	output, err := kubectl(ctx, kubeconfigPath, "get", "daemonset",
		"monitoring-agent", "-n", "monitoring",
		"-o", "jsonpath={.status.numberReady}")
	if err != nil {
		return fmt.Errorf("DaemonSet not found: %w", err)
	}

	numberReady := strings.TrimSpace(output)
	if numberReady == "0" || numberReady == "" {
		// Check if nodeSelector was removed or nodes were labeled
		dsYaml, err := kubectl(ctx, kubeconfigPath, "get", "daemonset",
			"monitoring-agent", "-n", "monitoring", "-o", "yaml")
		if err != nil {
			return fmt.Errorf("getting DaemonSet YAML: %w", err)
		}

		// If nodeSelector is still there, check if nodes have been labeled
		if strings.Contains(dsYaml, "node-type: monitoring") {
			// Check if any nodes have the label
			nodes, err := kubectl(ctx, kubeconfigPath, "get", "nodes",
				"-l", "node-type=monitoring", "-o", "name")
			if err != nil || strings.TrimSpace(nodes) == "" {
				return fmt.Errorf("DaemonSet still has no pods ready and nodes are not labeled")
			}
		}

		// If we get here, configuration looks good but pods might still be starting
		return fmt.Errorf("DaemonSet configuration fixed but pods not ready yet")
	}

	// Check actual pod status
	pods, err := kubectl(ctx, kubeconfigPath, "get", "pods",
		"-n", "monitoring", "-l", "app=monitoring-agent",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("checking DaemonSet pods: %w", err)
	}

	if !strings.Contains(pods, "Running") {
		return fmt.Errorf("DaemonSet pods exist but are not running yet")
	}

	return nil
}

func (l *DaemonSetNotScheduledLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check the DaemonSet status",
			Command:     "kubectl get daemonset -n monitoring",
			Notes:       "Should show 0 desired or 0 ready pods",
		},
		{
			Description: "Describe the DaemonSet",
			Command:     "kubectl describe daemonset monitoring-agent -n monitoring",
			Notes:       "Look for scheduling events and warnings",
		},
		{
			Description: "Check the DaemonSet YAML",
			Command:     "kubectl get daemonset monitoring-agent -n monitoring -o yaml | grep -A 5 nodeSelector",
			Notes:       "Look for node selector requirements",
		},
		{
			Description: "Check current node labels",
			Command:     "kubectl get nodes --show-labels",
			Notes:       "See what labels the nodes actually have",
		},
		{
			Description: "Identify the mismatch",
			Notes:       "The DaemonSet requires 'node-type: monitoring' but nodes don't have this label",
		},
		{
			Description: "Fix option 1: Remove the nodeSelector",
			Command:     "kubectl patch daemonset monitoring-agent -n monitoring --type json -p='[{\"op\": \"remove\", \"path\": \"/spec/template/spec/nodeSelector\"}]'",
			Notes:       "This allows the DaemonSet to run on all nodes",
		},
		{
			Description: "Fix option 2: Label the nodes (alternative)",
			Command:     "kubectl label nodes --all node-type=monitoring",
			Notes:       "This adds the required label to all nodes",
		},
		{
			Description: "Verify the DaemonSet is now scheduling pods",
			Command:     "kubectl get daemonset -n monitoring -w",
			Notes:       "Watch as the DaemonSet schedules pods on nodes",
		},
		{
			Description: "Check the pods are running",
			Command:     "kubectl get pods -n monitoring -l app=monitoring-agent",
			Notes:       "All pods should be in Running state",
		},
	}
}
