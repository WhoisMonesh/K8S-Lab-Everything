package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&NodeAffinityMismatchLab{})
}

type NodeAffinityMismatchLab struct {
	BaseLab
}

func (l *NodeAffinityMismatchLab) ID() string {
	return "node_affinity_mismatch"
}

func (l *NodeAffinityMismatchLab) Title() string {
	return "Pods Stuck By Impossible Node Affinity"
}

func (l *NodeAffinityMismatchLab) Category() Category {
	return CategoryScheduling
}

func (l *NodeAffinityMismatchLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *NodeAffinityMismatchLab) Description() string {
	return `An ETL deployment named 'etl-job' cannot place its single replica anywhere.
The pod stays Pending indefinitely.

Someone hardened the deployment's node affinity rules, and no node in this
cluster satisfies them anymore.

Your task: Adjust the affinity rules so the pod schedules onto a real node -
keep using node affinity, do not simply strip all constraints blindly.`
}

func (l *NodeAffinityMismatchLab) Hints() []string {
	return []string{
		"The unschedulable event names the exact requirement no node satisfies",
		"Dump the real labels present in the cluster: kubectl get nodes --show-labels",
		"requiredDuringScheduling... must match at least one node or the pod never places",
		"A safe target is a built-in label every Linux node carries, such as kubernetes.io/os",
	}
}

func (l *NodeAffinityMismatchLab) EstimatedTime() int {
	return 25
}

func (l *NodeAffinityMismatchLab) Tags() []string {
	return []string{"affinity", "node-selector", "labels", "scheduling", "troubleshooting"}
}

func (l *NodeAffinityMismatchLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodeAffinityMismatchLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: etl-job
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: etl-job
  template:
    metadata:
      labels:
        app: etl-job
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: storage-tier
                operator: In
                values:
                - nvme-ultra
              - key: region
                operator: In
                values:
                - eu-central-9
      containers:
      - name: etl
        image: busybox:1.36
        command: ["sh", "-c", "while true; do echo extracting; sleep 20; done"]
`

	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating broken deployment: %w", err)
	}
	return nil
}

func (l *NodeAffinityMismatchLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=etl-job",
		"-o", "jsonpath={.items[*].status.phase}")
	for _, phase := range splitFields(output) {
		if phase == "Running" {
			return fmt.Errorf("expected pod to be Pending")
		}
	}
	return nil
}

func (l *NodeAffinityMismatchLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "etl-job",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}
	if output != "1" {
		return fmt.Errorf("deployment not ready yet (ready replicas: %s, expected: 1)", output)
	}

	affinity, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "etl-job",
		"-o", "jsonpath={.spec.template.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution}")
	if err != nil {
		return fmt.Errorf("failed to read affinity: %w", err)
	}
	if affinity == "" {
		return fmt.Errorf("all node affinity was removed - fix it by matching real node labels instead")
	}
	return nil
}

func (l *NodeAffinityMismatchLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Read the scheduling failure",
			Command:     "kubectl describe pod -l app=etl-job | grep -A 8 Events",
			Notes:       "The message lists the nodeSelectorTerms that matched zero nodes",
		},
		{
			Description: "Inventory the labels that actually exist",
			Command:     "kubectl get nodes --show-labels",
			Notes:       "There is no storage-tier or region label anywhere in this cluster",
		},
		{
			Description: "Rewrite the affinity against a real label",
			Command:     "kubectl edit deploy etl-job",
			Notes:       "Replace the two matchExpressions with one targeting kubernetes.io/os: linux (or any label you saw in step 2)",
		},
		{
			Description: "Alternative: label a node to satisfy the rule",
			Command:     "kubectl label node <node-name> storage-tier=nvme-ultra region=eu-central-9",
			Notes:       "In production prefer aligning rules to existing infrastructure labels",
		},
		{
			Description: "Confirm the pod placed",
			Command:     "kubectl rollout status deploy/etl-job && kubectl get pods -l app=etl-job -o wide",
			Notes:       "The replica should be Running on the node you targeted",
		},
	}
}
