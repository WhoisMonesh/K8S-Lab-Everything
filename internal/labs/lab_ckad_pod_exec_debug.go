package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADPodExecDebugLab{})
}

type CKADPodExecDebugLab struct {
	BaseLab
}

func (l *CKADPodExecDebugLab) ID() string             { return "ckad_pod_exec_debug" }
func (l *CKADPodExecDebugLab) Title() string          { return "Debug Using kubectl exec" }
func (l *CKADPodExecDebugLab) Category() Category     { return CategoryAppObservability }
func (l *CKADPodExecDebugLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADPodExecDebugLab) Cert() Cert             { return CertCKAD }
func (l *CKADPodExecDebugLab) DomainWeight() int      { return 15 }
func (l *CKADPodExecDebugLab) EstimatedTime() int     { return 10 }
func (l *CKADPodExecDebugLab) Tags() []string {
	return []string{"exec", "debugging", "interactive"}
}

func (l *CKADPodExecDebugLab) Description() string {
	return `A pod is running but you need to investigate the environment inside the
container. Use kubectl exec to run commands inside the pod.

Your task: Execute commands inside the running pod to debug the issue.`
}

func (l *CKADPodExecDebugLab) Hints() []string {
	return []string{
		"Use kubectl exec to run commands in the pod",
		"Use -it for interactive terminal sessions",
		"Check environment variables, network, and filesystem",
	}
}

func (l *CKADPodExecDebugLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADPodExecDebugLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: debug-target
  labels:
    app: debug-target
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'while true; do sleep 3600; done']`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADPodExecDebugLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "debug-target", "--", "echo", "debug-ok")
	if err != nil {
		return fmt.Errorf("failed to exec into pod: %w", err)
	}
	if !strings.Contains(output, "debug-ok") {
		return fmt.Errorf("exec output unexpected: %s", output)
	}
	return nil
}

func (l *CKADPodExecDebugLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Get pod name", Command: "kubectl get pods -l app=debug-target"},
		{Description: "Execute command", Command: "kubectl exec debug-target -- ls /"},
		{Description: "Interactive session", Command: "kubectl exec -it debug-target -- sh"},
	}
}
