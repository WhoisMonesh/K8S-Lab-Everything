package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NodePIDPressureLab{})
}

type NodePIDPressureLab struct {
	BaseLab
}

func (l *NodePIDPressureLab) ID() string {
	return "node_pid_pressure"
}

func (l *NodePIDPressureLab) Title() string {
	return "Node PIDPressure Troubleshooting"
}

func (l *NodePIDPressureLab) Category() Category {
	return CategoryTroubleshooting
}

func (l *NodePIDPressureLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *NodePIDPressureLab) Description() string {
	return `A node is experiencing PIDPressure due to too many processes.
New pods cannot be scheduled on this node and existing pods may
be evicted.

Your task: Diagnose the PIDPressure condition and free up resources
by reducing the process count on the affected node.`
}

func (l *NodePIDPressureLab) Hints() []string {
	return []string{
		"Check node conditions for PIDPressure",
		"Look for pods with high process counts",
		"Check if any pods are spawning excessive processes",
	}
}

func (l *NodePIDPressureLab) EstimatedTime() int {
	return 20
}

func (l *NodePIDPressureLab) Tags() []string {
	return []string{"pid-pressure", "node", "troubleshooting", "resources", "eviction"}
}

func (l *NodePIDPressureLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NodePIDPressureLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: process-hog
  namespace: default
spec:
  replicas: 5
  selector:
    matchLabels:
      app: process-hog
  template:
    metadata:
      labels:
        app: process-hog
    spec:
      containers:
      - name: hog
        image: busybox:1.36
        command: ['sh', '-c', 'for i in $(seq 1 50); do sh -c "while true; do sleep 1000; done" & done; while true; do sleep 1000; done']
        resources:
          limits:
            memory: 32Mi
            cpu: 50m
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *NodePIDPressureLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(20 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=process-hog",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return nil
	}

	if strings.Contains(output, "Running") {
		return nil
	}

	return fmt.Errorf("pods not running")
}

func (l *NodePIDPressureLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=process-hog",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("checking pods: %w", err)
	}

	phases := strings.Fields(output)
	running := 0
	for _, p := range phases {
		if p == "Running" {
			running++
		}
	}

	if running > 0 {
		return fmt.Errorf("process-hog pods still running")
	}

	return nil
}

func (l *NodePIDPressureLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check node conditions",
			Command:     "kubectl get nodes -o custom-columns=NAME:.metadata.name,CONDITIONS:.status.conditions[*].type",
			Notes:       "Look for PIDPressure condition",
		},
		{
			Description: "Find resource-heavy pods",
			Command:     "kubectl top pods --sort-by=cpu",
			Notes:       "Identify pods consuming most resources",
		},
		{
			Description: "Fix: Scale down or delete the offending deployment",
			Command:     "kubectl scale deploy process-hog --replicas=0",
			Notes:       "Reduce process count to free PID resources",
		},
		{
			Description: "Verify PIDPressure is cleared",
			Command:     "kubectl get nodes | grep PIDPressure",
			Notes:       "PIDPressure condition should be False",
		},
	}
}
