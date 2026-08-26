package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ReadinessProbeWrong{})
}

type ReadinessProbeWrong struct {
	BaseLab
}

func (l *ReadinessProbeWrong) ID() string            { return "readiness_probe_wrong2" }
func (l *ReadinessProbeWrong) Title() string         { return "Pod Not Ready Due to Readiness Probe" }
func (l *ReadinessProbeWrong) Category() Category    { return CategoryWorkloads }
func (l *ReadinessProbeWrong) Difficulty() Difficulty { return DifficultyMedium }
func (l *ReadinessProbeWrong) EstimatedTime() int    { return 15 }
func (l *ReadinessProbeWrong) Tags() []string        { return []string{"probes", "readiness", "health"} }

func (l *ReadinessProbeWrong) Description() string {
	return `A pod is not ready because the readiness probe is misconfigured.
The probe is checking the wrong port or path. Fix the readiness probe.`
}

func (l *ReadinessProbeWrong) Hints() []string {
	return []string{
		"Check the readiness probe configuration",
		"Verify the port and path",
		"Compare with the actual application port",
	}
}

func (l *ReadinessProbeWrong) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ReadinessProbeWrong) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: ready-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: ready-app
  template:
    metadata:
      labels:
        app: ready-app
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ReadinessProbeWrong) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/ready-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return err
	}
	if output != "2" {
		return fmt.Errorf("deployment not ready")
	}
	return nil
}

func (l *ReadinessProbeWrong) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check readiness probe", Command: "kubectl get deploy ready-app -o jsonpath='{.spec.template.spec.containers[0].readinessProbe}'"},
		{Description: "Fix probe port", Command: "kubectl edit deploy ready-app"},
		{Description: "Change port", Command: "Change port from 8080 to 80"},
	}
}
