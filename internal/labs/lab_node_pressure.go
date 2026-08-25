package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NodePressureLab{})
}

type NodePressureLab struct {
	BaseLab
}

func (l *NodePressureLab) ID() string {
	return "node_pressure"
}

func (l *NodePressureLab) Title() string {
	return "Node Under Disk/Memory Pressure"
}

func (l *NodePressureLab) Category() Category {
	return CategoryControlPlane
}

func (l *NodePressureLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *NodePressureLab) Description() string {
	return `A node is experiencing memory pressure. New pods cannot be scheduled on this node,
and existing pods may be evicted. The kubelet is reporting pressure conditions.

Your task: Free up resources on the node to clear the pressure condition.`
}

func (l *NodePressureLab) Hints() []string {
	return []string{
		"Check node conditions",
		"Look at allocatable vs allocated resources",
		"Check for pods consuming excessive resources",
		"Look at disk usage on the node",
	}
}

func (l *NodePressureLab) EstimatedTime() int {
	return 25
}

func (l *NodePressureLab) Tags() []string {
	return []string{"node", "pressure", "eviction", "resources", "scheduling"}
}

func (l *NodePressureLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodePressureLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("getting control plane node: %w", err)
	}

	// Fill up disk space to trigger disk pressure
	_, err = dockerExec(ctx, nodeName, "sh", "-c",
		`dd if=/dev/zero of=/var/lib/kubelet/disk-hog bs=1M count=500 2>/dev/null; echo "disk hog created"`)
	if err != nil {
		return fmt.Errorf("creating disk pressure: %w", err)
	}

	// Create a pod that consumes a lot of memory to trigger memory pressure
	memoryHog := `apiVersion: v1
kind: Pod
metadata:
  name: memory-hog
  namespace: default
spec:
  nodeName: ` + nodeName + `
  containers:
  - name: hog
    image: busybox:1.28
    command: ['sh', '-c', 'dd if=/dev/urandom of=/dev/null bs=1M &
    while true; do sleep 3600; done']
    resources:
      requests:
        memory: "512Mi"
      limits:
        memory: "512Mi"
`
	if err := kubectlApply(ctx, kubeconfigPath, memoryHog); err != nil {
		return fmt.Errorf("creating memory hog pod: %w", err)
	}

	return nil
}

func (l *NodePressureLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(20 * time.Second)
	return nil
}

func (l *NodePressureLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check node conditions for pressure
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes",
		"-o", "jsonpath={.items[*].status.conditions[?(@.type=='DiskPressure')].status}")
	if err != nil {
		return fmt.Errorf("failed to check nodes: %w", err)
	}

	if strings.Contains(output, "True") {
		return fmt.Errorf("node still has DiskPressure")
	}

	// Check memory pressure
	output, err = kubectl(ctx, kubeconfigPath, "get", "nodes",
		"-o", "jsonpath={.items[*].status.conditions[?(@.type=='MemoryPressure')].status}")
	if err != nil {
		return fmt.Errorf("failed to check memory pressure: %w", err)
	}

	if strings.Contains(output, "True") {
		return fmt.Errorf("node still has MemoryPressure")
	}

	return nil
}

func (l *NodePressureLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check node conditions",
			Command:     "kubectl get nodes -o wide",
			Notes:       "Look at the STATUS column for pressure indicators",
		},
		{
			Description: "Describe the node",
			Command:     "kubectl describe node <node-name> | grep -A 10 Conditions",
			Notes:       "Look for DiskPressure or MemoryPressure set to True",
		},
		{
			Description: "Check disk usage on node",
			Command:     "docker exec <node-name> df -h /var/lib/kubelet",
			Notes:       "Disk usage will be very high",
		},
		{
			Description: "Remove the disk hog file",
			Command:     "docker exec <node-name> rm /var/lib/kubelet/disk-hog",
			Notes:       "Delete the file that's consuming disk space",
		},
		{
			Description: "Delete the memory hog pod",
			Command:     "kubectl delete pod memory-hog",
			Notes:       "Remove the pod consuming excessive memory",
		},
		{
			Description: "Wait for node conditions to clear",
			Command:     "kubectl get nodes -w",
			Notes:       "Wait for the pressure conditions to change to False",
		},
		{
			Description: "Verify node is healthy",
			Command:     "kubectl describe node <node-name> | grep -A 10 Conditions",
			Notes:       "All conditions should be False/Ready",
		},
	}
}
