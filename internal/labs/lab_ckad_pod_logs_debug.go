package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADPodLogsDebugLab{})
}

type CKADPodLogsDebugLab struct {
	BaseLab
}

func (l *CKADPodLogsDebugLab) ID() string             { return "ckad_pod_logs_debug" }
func (l *CKADPodLogsDebugLab) Title() string          { return "Debug Using Pod Logs" }
func (l *CKADPodLogsDebugLab) Category() Category     { return CategoryAppObservability }
func (l *CKADPodLogsDebugLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADPodLogsDebugLab) Cert() Cert             { return CertCKAD }
func (l *CKADPodLogsDebugLab) DomainWeight() int      { return 15 }
func (l *CKADPodLogsDebugLab) EstimatedTime() int     { return 10 }
func (l *CKADPodLogsDebugLab) Tags() []string {
	return []string{"logs", "debugging", "troubleshooting"}
}

func (l *CKADPodLogsDebugLab) Description() string {
	return `A pod is running but the application is failing. You need to investigate
the issue by examining the pod logs.

Your task: Retrieve the logs from the failing pod to identify the issue.`
}

func (l *CKADPodLogsDebugLab) Hints() []string {
	return []string{
		"Use kubectl logs to view pod logs",
		"Use --previous to see logs from crashed containers",
		"Use -f to follow log output in real-time",
	}
}

func (l *CKADPodLogsDebugLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADPodLogsDebugLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: failing-app
  labels:
    app: failing-app
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'echo "Error: connection refused" >&2 && exit 1']`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADPodLogsDebugLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "logs", "failing-app", "--tail=10")
	if err != nil {
		return fmt.Errorf("failed to get logs: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no logs found")
	}
	return nil
}

func (l *CKADPodLogsDebugLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod status", Command: "kubectl get pod failing-app"},
		{Description: "View logs", Command: "kubectl logs failing-app"},
		{Description: "Check previous container logs", Command: "kubectl logs failing-app --previous"},
	}
}
