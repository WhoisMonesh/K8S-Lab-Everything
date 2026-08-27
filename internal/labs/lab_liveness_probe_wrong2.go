package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&LivenessProbeWrong{})
}

type LivenessProbeWrong struct {
	BaseLab
}

func (l *LivenessProbeWrong) ID() string             { return "liveness_probe_wrong2" }
func (l *LivenessProbeWrong) Title() string          { return "Pod Failing Liveness Probe" }
func (l *LivenessProbeWrong) Category() Category     { return CategoryWorkloads }
func (l *LivenessProbeWrong) Difficulty() Difficulty { return DifficultyMedium }
func (l *LivenessProbeWrong) EstimatedTime() int     { return 15 }
func (l *LivenessProbeWrong) Tags() []string         { return []string{"probes", "liveness", "restarts"} }

func (l *LivenessProbeWrong) Description() string {
	return `A pod is being restarted because the liveness probe is failing.
The probe is checking the wrong path. Fix the liveness probe configuration.`
}

func (l *LivenessProbeWrong) Hints() []string {
	return []string{
		"Check the liveness probe configuration",
		"Verify the HTTP path",
		"Test the path manually with curl",
	}
}

func (l *LivenessProbeWrong) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *LivenessProbeWrong) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: live-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: live-app
  template:
    metadata:
      labels:
        app: live-app
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
        livenessProbe:
          httpGet:
            path: /health
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *LivenessProbeWrong) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/live-app",
		"-o", "jsonpath={.spec.template.spec.containers[0].livenessProbe.httpGet.path}")
	if err != nil {
		return err
	}
	if output == "/health" {
		return fmt.Errorf("liveness probe path still wrong")
	}
	return nil
}

func (l *LivenessProbeWrong) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check liveness probe", Command: "kubectl get deploy live-app -o jsonpath='{.spec.template.spec.containers[0].livenessProbe}'"},
		{Description: "Fix probe path", Command: "kubectl edit deploy live-app"},
		{Description: "Change path", Command: "Change path from /health to /"},
	}
}
