package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADLivenessProbeExecLab{})
}

type CKADLivenessProbeExecLab struct {
	BaseLab
}

func (l *CKADLivenessProbeExecLab) ID() string             { return "ckad_liveness_probe_exec" }
func (l *CKADLivenessProbeExecLab) Title() string          { return "Configure Exec-based Liveness Probe" }
func (l *CKADLivenessProbeExecLab) Category() Category     { return CategoryAppObservability }
func (l *CKADLivenessProbeExecLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADLivenessProbeExecLab) Cert() Cert             { return CertCKAD }
func (l *CKADLivenessProbeExecLab) DomainWeight() int      { return 15 }
func (l *CKADLivenessProbeExecLab) EstimatedTime() int     { return 15 }
func (l *CKADLivenessProbeExecLab) Tags() []string {
	return []string{"liveness-probe", "exec", "health-check"}
}

func (l *CKADLivenessProbeExecLab) Description() string {
	return `A pod is running but the application may become unresponsive. Add a
liveness probe that checks if the application is healthy by running
a command inside the container.

Your task: Add an exec-based liveness probe to the pod.`
}

func (l *CKADLivenessProbeExecLab) Hints() []string {
	return []string{
		"Use livenessProbe with exec action",
		"The command should return 0 for healthy state",
		"Set appropriate initialDelaySeconds and periodSeconds",
	}
}

func (l *CKADLivenessProbeExecLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADLivenessProbeExecLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: liveness-app
  labels:
    app: liveness-app
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'touch /tmp/healthy && sleep 3600']
    volumeMounts:
    - name: tmp
      mountPath: /tmp
  volumes:
  - name: tmp
    emptyDir: {}`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADLivenessProbeExecLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "liveness-app",
		"-o", "jsonpath={.spec.containers[0].livenessProbe.exec.command[*]}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no liveness probe configured")
	}
	return nil
}

func (l *CKADLivenessProbeExecLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod", Command: "kubectl edit pod liveness-app"},
		{Description: "Add liveness probe", Command: "Add livenessProbe with exec command ['cat', '/tmp/healthy']"},
		{Description: "Verify probe", Command: "kubectl get pod liveness-app -o yaml | grep -A 5 livenessProbe"},
	}
}
