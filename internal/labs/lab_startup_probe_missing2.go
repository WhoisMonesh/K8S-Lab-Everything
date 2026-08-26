package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&StartupProbeMissing{})
}

type StartupProbeMissing struct {
	BaseLab
}

func (l *StartupProbeMissing) ID() string            { return "startup_probe_missing2" }
func (l *StartupProbeMissing) Title() string         { return "Liveness Probe Kills Slow-Starting App" }
func (l *StartupProbeMissing) Category() Category    { return CategoryWorkloads }
func (l *StartupProbeMissing) Difficulty() Difficulty { return DifficultyMedium }
func (l *StartupProbeMissing) EstimatedTime() int    { return 20 }
func (l *StartupProbeMissing) Tags() []string        { return []string{"probes", "startup", "liveness"} }

func (l *StartupProbeMissing) Description() string {
	return `A slow-starting application is being killed by the liveness probe before it can start.
Add a startup probe to allow the application time to start.`
}

func (l *StartupProbeMissing) Hints() []string {
	return []string{
		"Check the liveness probe configuration",
		"Look at the application startup time",
		"Add a startup probe with a longer initial delay",
	}
}

func (l *StartupProbeMissing) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *StartupProbeMissing) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: slow-start
spec:
  replicas: 1
  selector:
    matchLabels:
      app: slow-start
  template:
    metadata:
      labels:
        app: slow-start
    spec:
      containers:
      - name: app
        image: nginx:alpine
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *StartupProbeMissing) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/slow-start",
		"-o", "jsonpath={.spec.template.spec.containers[0].startupProbe}")
	if err != nil {
		return err
	}
	if output == "" || output == "null" {
		return fmt.Errorf("startup probe not configured")
	}
	return nil
}

func (l *StartupProbeMissing) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check probes", Command: "kubectl get deploy slow-start -o jsonpath='{.spec.template.spec.containers[0].livenessProbe}'"},
		{Description: "Add startup probe", Command: "kubectl edit deploy slow-start"},
		{Description: "Add startupProbe config", Command: "Add startupProbe with httpGet path / port 80, failureThreshold 30, periodSeconds 10"},
	}
}
