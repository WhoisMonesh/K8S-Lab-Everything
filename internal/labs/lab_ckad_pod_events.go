package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADPodEventsLab{})
}

type CKADPodEventsLab struct {
	BaseLab
}

func (l *CKADPodEventsLab) ID() string             { return "ckad_pod_events" }
func (l *CKADPodEventsLab) Title() string          { return "Debug Using Pod Events" }
func (l *CKADPodEventsLab) Category() Category     { return CategoryAppObservability }
func (l *CKADPodEventsLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADPodEventsLab) Cert() Cert             { return CertCKAD }
func (l *CKADPodEventsLab) DomainWeight() int      { return 15 }
func (l *CKADPodEventsLab) EstimatedTime() int     { return 10 }
func (l *CKADPodEventsLab) Tags() []string {
	return []string{"events", "debugging", "troubleshooting"}
}

func (l *CKADPodEventsLab) Description() string {
	return `A pod is stuck in a pending state. Use Kubernetes events to understand
why the pod cannot be scheduled.

Your task: Check the pod events to identify the scheduling issue.`
}

func (l *CKADPodEventsLab) Hints() []string {
	return []string{
		"Use kubectl describe pod to see events",
		"Use kubectl get events to see cluster events",
		"Filter events by namespace or resource",
	}
}

func (l *CKADPodEventsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADPodEventsLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: pending-pod
  labels:
    app: pending-pod
spec:
  nodeSelector:
    disktype: ssd
  containers:
  - name: app
    image: nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADPodEventsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "describe", "pod", "pending-pod")
	if err != nil {
		return fmt.Errorf("failed to describe pod: %w", err)
	}
	if strings.Contains(output, "Events") || strings.Contains(output, "node(s) didn't match") {
		return nil
	}
	return fmt.Errorf("events not found in describe output")
}

func (l *CKADPodEventsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod status", Command: "kubectl get pod pending-pod"},
		{Description: "Describe pod for events", Command: "kubectl describe pod pending-pod"},
		{Description: "Get events", Command: "kubectl get events --field-selector involvedObject.name=pending-pod"},
	}
}
