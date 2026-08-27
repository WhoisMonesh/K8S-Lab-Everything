package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADStartupProbeLab{})
}

type CKADStartupProbeLab struct {
	BaseLab
}

func (l *CKADStartupProbeLab) ID() string             { return "ckad_startup_probe" }
func (l *CKADStartupProbeLab) Title() string          { return "Configure Startup Probe" }
func (l *CKADStartupProbeLab) Category() Category     { return CategoryAppObservability }
func (l *CKADStartupProbeLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADStartupProbeLab) Cert() Cert             { return CertCKAD }
func (l *CKADStartupProbeLab) DomainWeight() int      { return 15 }
func (l *CKADStartupProbeLab) EstimatedTime() int     { return 15 }
func (l *CKADStartupProbeLab) Tags() []string {
	return []string{"startup-probe", "health-check", "slow-start"}
}

func (l *CKADStartupProbeLab) Description() string {
	return `A slow-starting application needs a startup probe to allow it time to
initialize before liveness and readiness probes begin checking.

Your task: Add a startup probe that gives the application 60 seconds to start.`
}

func (l *CKADStartupProbeLab) Hints() []string {
	return []string{
		"Startup probes run before liveness and readiness probes",
		"Set failureThreshold high enough for slow startups",
		"Use httpGet or exec action for the probe",
	}
}

func (l *CKADStartupProbeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADStartupProbeLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: slow-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: slow-app
  template:
    metadata:
      labels:
        app: slow-app
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80
        livenessProbe:
          httpGet:
            path: /healthz
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 10`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *CKADStartupProbeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "slow-app",
		"-o", "jsonpath={.spec.template.spec.containers[0].startupProbe}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no startup probe configured")
	}
	return nil
}

func (l *CKADStartupProbeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit deployment", Command: "kubectl edit deployment slow-app"},
		{Description: "Add startup probe", Command: "Add startupProbe with httpGet path /healthz port 80 and failureThreshold: 30"},
		{Description: "Verify probe", Command: "kubectl get deployment slow-app -o yaml | grep -A 5 startupProbe"},
	}
}
