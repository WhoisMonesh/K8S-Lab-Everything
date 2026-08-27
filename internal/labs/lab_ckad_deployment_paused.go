package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADDeploymentPausedLab{})
}

type CKADDeploymentPausedLab struct {
	BaseLab
}

func (l *CKADDeploymentPausedLab) ID() string             { return "ckad_deployment_paused" }
func (l *CKADDeploymentPausedLab) Title() string          { return "Handle Paused Deployment" }
func (l *CKADDeploymentPausedLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADDeploymentPausedLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADDeploymentPausedLab) Cert() Cert             { return CertCKAD }
func (l *CKADDeploymentPausedLab) DomainWeight() int      { return 20 }
func (l *CKADDeploymentPausedLab) EstimatedTime() int     { return 10 }
func (l *CKADDeploymentPausedLab) Tags() []string {
	return []string{"paused", "deployment", "rollout"}
}

func (l *CKADDeploymentPausedLab) Description() string {
	return `A deployment is paused and updates are not being applied. New pods are
not being created when the deployment spec is changed.

Your task: Resume the paused deployment so updates can proceed.`
}

func (l *CKADDeploymentPausedLab) Hints() []string {
	return []string{
		"Check if the deployment is paused with kubectl get deployment",
		"Use kubectl rollout resume to unpause",
		"Verify the rollout continues after resuming",
	}
}

func (l *CKADDeploymentPausedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADDeploymentPausedLab) Break(ctx context.Context, kubeconfigPath string) error {
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

	pause := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
spec:
  paused: true
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
	return kubectlApply(ctx, kubeconfigPath, pause)
}

func (l *CKADDeploymentPausedLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp",
		"-o", "jsonpath={.spec.paused}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if strings.TrimSpace(output) == "true" {
		return fmt.Errorf("deployment is still paused")
	}
	return nil
}

func (l *CKADDeploymentPausedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check deployment status", Command: "kubectl get deployment webapp"},
		{Description: "Resume deployment", Command: "kubectl rollout resume deployment/webapp"},
		{Description: "Verify rollout", Command: "kubectl rollout status deployment/webapp"},
	}
}
