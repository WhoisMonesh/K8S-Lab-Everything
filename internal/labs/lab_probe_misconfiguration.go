package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ProbeMisconfigurationLab{})
}

type ProbeMisconfigurationLab struct {
	BaseLab
}

func (l *ProbeMisconfigurationLab) ID() string {
	return "probe_misconfiguration"
}

func (l *ProbeMisconfigurationLab) Title() string {
	return "Liveness/Readiness Probe Misconfiguration"
}

func (l *ProbeMisconfigurationLab) Category() Category {
	return CategoryAppObservability
}

func (l *ProbeMisconfigurationLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ProbeMisconfigurationLab) Description() string {
	return `A deployment has misconfigured liveness and readiness probes causing
pods to crash loop or never become ready.

Your task: Fix the probe configuration so pods start correctly and become ready.`
}

func (l *ProbeMisconfigurationLab) Hints() []string {
	return []string{
		"Check the pod status and describe events",
		"HTTP probes need a valid path and port",
		"Initial delay might be too short for the app to start",
	}
}

func (l *ProbeMisconfigurationLab) EstimatedTime() int {
	return 20
}

func (l *ProbeMisconfigurationLab) Tags() []string {
	return []string{"probes", "liveness", "readiness", "troubleshooting", "workloads"}
}

func (l *ProbeMisconfigurationLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ProbeMisconfigurationLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: probe-test
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: probe-test
  template:
    metadata:
      labels:
        app: probe-test
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo running; sleep 15; done']
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /healthz
            port: 9999
          initialDelaySeconds: 1
          periodSeconds: 5
        readinessProbe:
          httpGet:
            path: /ready
            port: 9999
          initialDelaySeconds: 1
          periodSeconds: 5
        resources:
          limits:
            memory: 32Mi
            cpu: 50m
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *ProbeMisconfigurationLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(20 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=probe-test",
		"-o", "jsonpath={.items[*].status.conditions[?(@.type==\"Ready\")].status}")
	if err != nil {
		return fmt.Errorf("checking pods: %w", err)
	}

	if strings.Contains(output, "False") || strings.Contains(output, "Unknown") {
		return nil
	}

	return fmt.Errorf("pods appear ready (expected not ready)")
}

func (l *ProbeMisconfigurationLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "probe-test",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("checking deployment: %w", err)
	}

	ready := strings.TrimSpace(output)
	if ready != "2" {
		return fmt.Errorf("deployment not ready (ready: %s, expected: 2)", ready)
	}

	return nil
}

func (l *ProbeMisconfigurationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pods -l app=probe-test",
			Notes:       "Pods should be in CrashLoopBackOff or not Ready",
		},
		{
			Description: "Describe pod for probe failures",
			Command:     "kubectl describe pod -l app=probe-test | grep -A 5 'Liveness\\|Readiness'",
			Notes:       "Probes are hitting port 9999 which doesn't exist",
		},
		{
			Description: "Fix: Update probe port to match container",
			Command:     `kubectl patch deploy probe-test --type='json' -p='[{"op":"replace","path":"/spec/template/spec/containers/0/livenessProbe/httpGet/port","value":8080},{"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/httpGet/port","value":8080}]'`,
			Notes:       "Port should match the containerPort (8080)",
		},
		{
			Description: "Verify pods are ready",
			Command:     "kubectl rollout status deploy/probe-test",
			Notes:       "All replicas should be ready",
		},
	}
}
