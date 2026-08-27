package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADReadinessProbeLab{})
}

type CKADReadinessProbeLab struct {
	BaseLab
}

func (l *CKADReadinessProbeLab) ID() string             { return "ckad_readiness_probe" }
func (l *CKADReadinessProbeLab) Title() string          { return "Configure Readiness Probe" }
func (l *CKADReadinessProbeLab) Category() Category     { return CategoryAppObservability }
func (l *CKADReadinessProbeLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADReadinessProbeLab) Cert() Cert             { return CertCKAD }
func (l *CKADReadinessProbeLab) DomainWeight() int      { return 15 }
func (l *CKADReadinessProbeLab) EstimatedTime() int     { return 15 }
func (l *CKADReadinessProbeLab) Tags() []string {
	return []string{"readiness-probe", "health-check", "service"}
}

func (l *CKADReadinessProbeLab) Description() string {
	return `A service is sending traffic to pods that aren't ready yet. Add a readiness
probe to ensure traffic only goes to healthy pods.

Your task: Add a readiness probe to the deployment that checks /ready.`
}

func (l *CKADReadinessProbeLab) Hints() []string {
	return []string{
		"Use readinessProbe with httpGet action",
		"Pods won't receive traffic until the probe succeeds",
		"Set appropriate initialDelaySeconds",
	}
}

func (l *CKADReadinessProbeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADReadinessProbeLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: webapp
        image: nginx:alpine
        ports:
        - containerPort: 80`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	svc := `apiVersion: v1
kind: Service
metadata:
  name: webapp
spec:
  selector:
    app: webapp
  ports:
  - port: 80
    targetPort: 80`
	return kubectlApply(ctx, kubeconfigPath, svc)
}

func (l *CKADReadinessProbeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp",
		"-o", "jsonpath={.spec.template.spec.containers[0].readinessProbe.httpGet.path}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no readiness probe configured")
	}
	return nil
}

func (l *CKADReadinessProbeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit deployment", Command: "kubectl edit deployment webapp"},
		{Description: "Add readiness probe", Command: "Add readinessProbe with httpGet path /ready port 80"},
		{Description: "Verify probe", Command: "kubectl get deployment webapp -o yaml | grep -A 5 readinessProbe"},
	}
}
