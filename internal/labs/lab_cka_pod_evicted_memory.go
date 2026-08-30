package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&PodEvictedMemoryLab{})
}

type PodEvictedMemoryLab struct {
	BaseLab
}

func (l *PodEvictedMemoryLab) ID() string { return "cka_pod_evicted_memory" }
func (l *PodEvictedMemoryLab) Title() string {
	return "Debug OOMKilled Pods"
}
func (l *PodEvictedMemoryLab) Category() Category     { return CategoryTroubleshooting }
func (l *PodEvictedMemoryLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PodEvictedMemoryLab) EstimatedTime() int     { return 20 }
func (l *PodEvictedMemoryLab) Tags() []string {
	return []string{"oomkilled", "memory", "resource-limits", "troubleshooting"}
}
func (l *PodEvictedMemoryLab) Cert() Cert        { return CertCKA }
func (l *PodEvictedMemoryLab) DomainWeight() int { return 30 }

func (l *PodEvictedMemoryLab) Description() string {
	return `A pod is being OOMKilled because its memory limit is too low. Investigate
the memory usage and increase the memory limit to prevent future kills.`
}

func (l *PodEvictedMemoryLab) Hints() []string {
	return []string{
		"Check pod last state for OOMKilled",
		"Review memory limits in container spec",
		"Increase memory limit appropriately",
	}
}

func (l *PodEvictedMemoryLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodEvictedMemoryLab) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: oom-ns
---
apiVersion: v1
kind: Pod
metadata:
  name: memory-hog
  namespace: oom-ns
spec:
  containers:
  - name: stress
    image: busybox:1.36
    command: ["/bin/sh", "-c", "while true; do cat /dev/zero > /dev/null; done"]
    resources:
      limits:
        memory: 32Mi
`
	if err := kubectlApply(ctx, kubeconfigPath, manifest); err != nil {
		return fmt.Errorf("creating broken pod: %w", err)
	}
	return nil
}

func (l *PodEvictedMemoryLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodEvictedMemoryLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "oom-ns",
		"-o", "jsonpath={.items[0].spec.containers[0].resources.limits.memory}")
	if err != nil {
		return err
	}
	if output == "64Mi" || output == "32Mi" {
		return fmt.Errorf("memory limit still too low")
	}
	return nil
}

func (l *PodEvictedMemoryLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod status", Command: "kubectl get pods -n oom-ns"},
		{Description: "Check last state", Command: "kubectl get pod -n oom-ns -o jsonpath='{.items[0].status.containerStatuses[0].lastState}'"},
		{Description: "Increase memory limit", Command: "kubectl set resources pod/memory-hog -n oom-ns --limits=memory=512Mi"},
		{Description: "Verify", Command: "kubectl get pods -n oom-ns"},
	}
}
