package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADDeploymentRollingUpdateLab{})
}

type CKADDeploymentRollingUpdateLab struct {
	BaseLab
}

func (l *CKADDeploymentRollingUpdateLab) ID() string             { return "ckad_deployment_rolling_update" }
func (l *CKADDeploymentRollingUpdateLab) Title() string          { return "Perform Rolling Update" }
func (l *CKADDeploymentRollingUpdateLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADDeploymentRollingUpdateLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADDeploymentRollingUpdateLab) Cert() Cert             { return CertCKAD }
func (l *CKADDeploymentRollingUpdateLab) DomainWeight() int      { return 20 }
func (l *CKADDeploymentRollingUpdateLab) EstimatedTime() int     { return 20 }
func (l *CKADDeploymentRollingUpdateLab) Tags() []string {
	return []string{"rolling-update", "deployment", "updates"}
}

func (l *CKADDeploymentRollingUpdateLab) Description() string {
	return `A deployment is running v1 of the application. You need to update it to v2
using a rolling update strategy with zero downtime.

Your task: Update the deployment image from nginx:1.21 to nginx:1.25
and verify the rolling update completes successfully.`
}

func (l *CKADDeploymentRollingUpdateLab) Hints() []string {
	return []string{
		"Use kubectl set image to update the container image",
		"Monitor the rollout status with kubectl rollout status",
		"Check rollout history with kubectl rollout history",
	}
}

func (l *CKADDeploymentRollingUpdateLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADDeploymentRollingUpdateLab) Break(ctx context.Context, kubeconfigPath string) error {
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
        image: nginx:1.21
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *CKADDeploymentRollingUpdateLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp",
		"-o", "jsonpath={.spec.template.spec.containers[0].image}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if !strings.Contains(output, "1.25") {
		return fmt.Errorf("image not updated to 1.25 (current: %s)", output)
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

func (l *CKADDeploymentRollingUpdateLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current image", Command: "kubectl get deployment webapp -o yaml | grep image"},
		{Description: "Update image", Command: "kubectl set image deployment/webapp webapp=nginx:1.25"},
		{Description: "Monitor rollout", Command: "kubectl rollout status deployment/webapp"},
		{Description: "Verify update", Command: "kubectl get deployment webapp -o yaml | grep image"},
	}
}
