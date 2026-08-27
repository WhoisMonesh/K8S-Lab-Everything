package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&LivenessProbeWrongLab{})
}

type LivenessProbeWrongLab struct {
	BaseLab
}

func (l *LivenessProbeWrongLab) ID() string {
	return "liveness_probe_wrong"
}

func (l *LivenessProbeWrongLab) Title() string {
	return "Pod Failing Liveness Probe"
}

func (l *LivenessProbeWrongLab) Category() Category {
	return CategoryWorkloads
}

func (l *LivenessProbeWrongLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *LivenessProbeWrongLab) Description() string {
	return `A deployment 'webapp' has pods continuously restarting due to liveness probe failures.
The pods start but get killed by kubelet because the liveness probe keeps failing.

Your task: Fix the liveness probe configuration so pods stay running.`
}

func (l *LivenessProbeWrongLab) Hints() []string {
	return []string{
		"Check the pod restart count",
		"Look at the liveness probe configuration",
		"The probe might be hitting the wrong port or path",
		"Check if the application is actually healthy",
	}
}

func (l *LivenessProbeWrongLab) EstimatedTime() int {
	return 15
}

func (l *LivenessProbeWrongLab) Tags() []string {
	return []string{"liveness-probe", "health-check", "restart", "deployment"}
}

func (l *LivenessProbeWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *LivenessProbeWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create deployment with wrong liveness probe (wrong port)
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: web
        image: nginx:alpine
        ports:
        - containerPort: 80
        livenessProbe:
          httpGet:
            path: /
            port: 9999
          initialDelaySeconds: 5
          periodSeconds: 10
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	return nil
}

func (l *LivenessProbeWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(20 * time.Second)
	return nil
}

func (l *LivenessProbeWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if pods have been restarting
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=webapp",
		"-o", "jsonpath={.items[*].status.restartCount}")
	if err != nil {
		return fmt.Errorf("failed to check pods: %w", err)
	}

	// After fix, restart count should stop increasing
	// Check if all pods are running
	output, err = kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=webapp",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod phases: %w", err)
	}

	for _, phase := range strings.Fields(output) {
		if phase != "Running" {
			return fmt.Errorf("not all pods are running (phases: %s)", output)
		}
	}

	// Check liveness probe port
	output, err = kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp",
		"-o", "jsonpath={.spec.template.spec.containers[0].livenessProbe.httpGet.port}")
	if err != nil {
		return fmt.Errorf("failed to check liveness probe: %w", err)
	}

	if strings.TrimSpace(output) != "80" {
		return fmt.Errorf("liveness probe port is wrong (got: %s, expected: 80)", output)
	}

	return nil
}

func (l *LivenessProbeWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check deployment status",
			Command:     "kubectl get deployment webapp",
			Notes:       "Notice the pods keep restarting",
		},
		{
			Description: "Check pod restart counts",
			Command:     "kubectl get pods -l app=webapp",
			Notes:       "RESTARTS column will show increasing numbers",
		},
		{
			Description: "Check pod events",
			Command:     "kubectl describe pod -l app=webapp | grep -A 10 Liveness",
			Notes:       "Look for 'Liveness probe failed' errors",
		},
		{
			Description: "Check the liveness probe configuration",
			Command:     "kubectl get deployment webapp -o yaml | grep -A 5 livenessProbe",
			Notes:       "The probe is targeting port 9999, but nginx listens on port 80",
		},
		{
			Description: "Fix the liveness probe port",
			Command:     "kubectl edit deployment webapp",
			Notes:       "Change the liveness probe port from 9999 to 80",
		},
		{
			Description: "Wait for rollout to complete",
			Command:     "kubectl rollout status deployment webapp",
			Notes:       "Wait for the new pods with fixed probe to start",
		},
		{
			Description: "Verify pods are stable",
			Command:     "kubectl get pods -l app=webapp",
			Notes:       "Pods should be Running and restart counts should stop increasing",
		},
	}
}
