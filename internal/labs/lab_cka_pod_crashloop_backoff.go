package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&PodCrashLoopBackoffLab{})
}

type PodCrashLoopBackoffLab struct {
	BaseLab
}

func (l *PodCrashLoopBackoffLab) ID() string { return "cka_pod_crashloop_backoff" }
func (l *PodCrashLoopBackoffLab) Title() string {
	return "Debug CrashLoopBackOff Pods"
}
func (l *PodCrashLoopBackoffLab) Category() Category     { return CategoryTroubleshooting }
func (l *PodCrashLoopBackoffLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PodCrashLoopBackoffLab) EstimatedTime() int     { return 20 }
func (l *PodCrashLoopBackoffLab) Tags() []string {
	return []string{"pod", "crashloopbackoff", "debugging", "troubleshooting"}
}
func (l *PodCrashLoopBackoffLab) Cert() Cert        { return CertCKA }
func (l *PodCrashLoopBackoffLab) DomainWeight() int { return 30 }

func (l *PodCrashLoopBackoffLab) Description() string {
	return `A pod is in CrashLoopBackOff state. Investigate the pod logs and events
to determine why the container is crashing. Fix the configuration issue
causing the crashes.`
}

func (l *PodCrashLoopBackoffLab) Hints() []string {
	return []string{
		"Check pod logs with kubectl logs",
		"Look at pod events with kubectl describe pod",
		"Verify the container command and environment variables",
	}
}

func (l *PodCrashLoopBackoffLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodCrashLoopBackoffLab) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: crash-ns
---
apiVersion: v1
kind: Pod
metadata:
  name: crash-app
  namespace: crash-ns
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ["/bin/sh", "-c", "exit 1"]
`
	if err := kubectlApply(ctx, kubeconfigPath, manifest); err != nil {
		return fmt.Errorf("creating broken pod: %w", err)
	}
	return nil
}

func (l *PodCrashLoopBackoffLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodCrashLoopBackoffLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "crash-ns",
		"-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("pod not in Running state")
	}
	return nil
}

func (l *PodCrashLoopBackoffLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod status", Command: "kubectl get pods -n crash-ns"},
		{Description: "Check logs", Command: "kubectl logs -n crash-ns <pod-name> --previous"},
		{Description: "Describe pod", Command: "kubectl describe pod -n crash-ns <pod-name>"},
		{Description: "Fix and apply", Command: "Fix the configuration and reapply"},
	}
}
