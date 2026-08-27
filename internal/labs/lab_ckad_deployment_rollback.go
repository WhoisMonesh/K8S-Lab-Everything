package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADDeploymentRollbackLab{})
}

type CKADDeploymentRollbackLab struct {
	BaseLab
}

func (l *CKADDeploymentRollbackLab) ID() string             { return "ckad_deployment_rollback" }
func (l *CKADDeploymentRollbackLab) Title() string          { return "Rollback Failed Deployment" }
func (l *CKADDeploymentRollbackLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADDeploymentRollbackLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADDeploymentRollbackLab) Cert() Cert             { return CertCKAD }
func (l *CKADDeploymentRollbackLab) DomainWeight() int      { return 20 }
func (l *CKADDeploymentRollbackLab) EstimatedTime() int     { return 15 }
func (l *CKADDeploymentRollbackLab) Tags() []string {
	return []string{"rollback", "deployment", "recovery"}
}

func (l *CKADDeploymentRollbackLab) Description() string {
	return `A deployment was updated to a broken version (nginx:nonexistent) and pods
are stuck in ImagePullBackOff.

Your task: Rollback the deployment to the previous working version.`
}

func (l *CKADDeploymentRollbackLab) Hints() []string {
	return []string{
		"Use kubectl rollout undo to rollback",
		"Check rollout history to see revisions",
		"Rollback to the previous revision",
	}
}

func (l *CKADDeploymentRollbackLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADDeploymentRollbackLab) Break(ctx context.Context, kubeconfigPath string) error {
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
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	broken := `apiVersion: apps/v1
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
        image: nginx:nonexistent
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, broken)
}

func (l *CKADDeploymentRollbackLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp",
		"-o", "jsonpath={.spec.template.spec.containers[0].image}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if strings.Contains(output, "nonexistent") {
		return fmt.Errorf("still on broken image: %s", output)
	}
	return nil
}

func (l *CKADDeploymentRollbackLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check rollout history", Command: "kubectl rollout history deployment/webapp"},
		{Description: "Rollback to previous version", Command: "kubectl rollout undo deployment/webapp"},
		{Description: "Verify rollback", Command: "kubectl rollout status deployment/webapp"},
		{Description: "Check current image", Command: "kubectl get deployment webapp -o yaml | grep image"},
	}
}
