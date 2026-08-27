package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodOverheadCgroupLab{})
}

type PodOverheadCgroupLab struct {
	BaseLab
}

func (l *PodOverheadCgroupLab) ID() string {
	return "pod_overhead_cgroup"
}

func (l *PodOverheadCgroupLab) Title() string {
	return "Pod Overhead Cgroup Mismatch"
}

func (l *PodOverheadCgroupLab) Category() Category {
	return CategoryScheduling
}

func (l *PodOverheadCgroupLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *PodOverheadCgroupLab) Description() string {
	return `A RuntimeClass 'gvisor' has overhead.podFixed set incorrectly. The
overhead is set to CPU: 2 cores but the node only has 1 core available.
Pods using this RuntimeClass cannot be scheduled.

Your task: Fix the RuntimeClass overhead to match node capacity.`
}

func (l *PodOverheadCgroupLab) Hints() []string {
	return []string{
		"Check the RuntimeClass overhead configuration",
		"Overhead values must not exceed node resources",
		"Reduce overhead to match available node capacity",
	}
}

func (l *PodOverheadCgroupLab) EstimatedTime() int {
	return 15
}

func (l *PodOverheadCgroupLab) Tags() []string {
	return []string{"runtimeclass", "overhead", "cgroup", "scheduling"}
}

func (l *PodOverheadCgroupLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodOverheadCgroupLab) Break(ctx context.Context, kubeconfigPath string) error {
	runtimeClass := `apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: gvisor
handler: gvisor
overhead:
  podFixed:
    cpu: "2"
    memory: "256Mi"
`
	if err := kubectlApply(ctx, kubeconfigPath, runtimeClass); err != nil {
		return fmt.Errorf("creating RuntimeClass: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: gvisor-pod
  namespace: default
spec:
  runtimeClassName: gvisor
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'sleep 3600']
    resources:
      requests:
        cpu: 100m
        memory: 64Mi
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodOverheadCgroupLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pod", "gvisor-pod",
		"-o", "jsonpath={.status.phase}")
	if strings.TrimSpace(output) == "Pending" {
		return nil
	}
	return nil
}

func (l *PodOverheadCgroupLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "runtimeclass", "gvisor",
		"-o", "jsonpath={.overhead.podFixed.cpu}")
	if err != nil {
		return fmt.Errorf("failed to check runtimeclass: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "2" {
		return fmt.Errorf("CPU overhead is still 2 cores")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "gvisor-pod",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *PodOverheadCgroupLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check RuntimeClass overhead",
			Command:     "kubectl get runtimeclass gvisor -o yaml | grep -A 5 overhead",
			Notes:       "CPU overhead is set to 2 cores which is too high",
		},
		{
			Description: "Check node resources",
			Command:     "kubectl describe nodes | grep -A 5 Allocatable",
			Notes:       "Nodes may not have 2 free CPU cores",
		},
		{
			Description: "Fix RuntimeClass overhead",
			Command:     "kubectl edit runtimeclass gvisor",
			Notes:       "Reduce CPU overhead to 100m or remove overhead",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod gvisor-pod",
			Notes:       "Pod should now be Running",
		},
	}
}
