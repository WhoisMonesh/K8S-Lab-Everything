package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADDeploymentProgressDeadlineLab{})
}

type CKADDeploymentProgressDeadlineLab struct {
	BaseLab
}

func (l *CKADDeploymentProgressDeadlineLab) ID() string {
	return "ckad_deployment_progress_deadline"
}

func (l *CKADDeploymentProgressDeadlineLab) Title() string {
	return "Fix Progress Deadline Exceeded"
}

func (l *CKADDeploymentProgressDeadlineLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADDeploymentProgressDeadlineLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADDeploymentProgressDeadlineLab) Cert() Cert             { return CertCKAD }
func (l *CKADDeploymentProgressDeadlineLab) DomainWeight() int      { return 20 }
func (l *CKADDeploymentProgressDeadlineLab) EstimatedTime() int     { return 20 }
func (l *CKADDeploymentProgressDeadlineLab) Tags() []string {
	return []string{"progress-deadline", "deployment", "troubleshooting"}
}

func (l *CKADDeploymentProgressDeadlineLab) Description() string {
	return `A deployment has exceeded its progress deadline because it's trying to pull
a non-existent image. The deployment is stuck and won't make progress.

Your task: Fix the deployment by correcting the image name so it can complete
the rollout.`
}

func (l *CKADDeploymentProgressDeadlineLab) Hints() []string {
	return []string{
		"Check the deployment events for the error",
		"The image name is incorrect",
		"Update the image to a valid one like nginx:alpine",
	}
}

func (l *CKADDeploymentProgressDeadlineLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADDeploymentProgressDeadlineLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
spec:
  replicas: 3
  progressDeadlineSeconds: 60
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
        image: nginx:nonexistent:latest
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *CKADDeploymentProgressDeadlineLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp",
		"-o", "jsonpath={.spec.template.spec.containers[0].image}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if strings.Contains(output, "nonexistent") {
		return fmt.Errorf("image still incorrect: %s", output)
	}

	ready, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to get ready replicas: %w", err)
	}
	if strings.TrimSpace(ready) == "0" || strings.TrimSpace(ready) == "" {
		return fmt.Errorf("deployment has no ready replicas")
	}
	return nil
}

func (l *CKADDeploymentProgressDeadlineLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check deployment status", Command: "kubectl get deployment webapp"},
		{Description: "View events", Command: "kubectl describe deployment webapp"},
		{Description: "Fix image", Command: "kubectl set image deployment/webapp webapp=nginx:alpine"},
		{Description: "Monitor rollout", Command: "kubectl rollout status deployment/webapp"},
	}
}
