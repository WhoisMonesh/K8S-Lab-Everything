package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADCanaryDeployLab{})
}

type CKADCanaryDeployLab struct {
	BaseLab
}

func (l *CKADCanaryDeployLab) ID() string             { return "ckad_canary_deploy" }
func (l *CKADCanaryDeployLab) Title() string          { return "Implement Canary Deployment" }
func (l *CKADCanaryDeployLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADCanaryDeployLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKADCanaryDeployLab) Cert() Cert             { return CertCKAD }
func (l *CKADCanaryDeployLab) DomainWeight() int      { return 20 }
func (l *CKADCanaryDeployLab) EstimatedTime() int     { return 25 }
func (l *CKADCanaryDeployLab) Tags() []string {
	return []string{"canary", "deployment", "strategy"}
}

func (l *CKADCanaryDeployLab) Description() string {
	return `Implement a canary deployment pattern where a small portion of traffic
is routed to the new version before full rollout.

Your task: Create a canary deployment with 1 replica alongside the main
deployment with 3 replicas, both selected by the same Service.`
}

func (l *CKADCanaryDeployLab) Hints() []string {
	return []string{
		"Canary uses a separate deployment with fewer replicas",
		"Both deployments share the same Service labels",
		"Adjust replica counts to control traffic percentage",
	}
}

func (l *CKADCanaryDeployLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADCanaryDeployLab) Break(ctx context.Context, kubeconfigPath string) error {
	main := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp-stable
spec:
  replicas: 3
  selector:
    matchLabels:
      app: webapp
      track: stable
  template:
    metadata:
      labels:
        app: webapp
        track: stable
    spec:
      containers:
      - name: webapp
        image: nginx:1.21
        ports:
        - containerPort: 80`
	if err := kubectlApply(ctx, kubeconfigPath, main); err != nil {
		return fmt.Errorf("creating main deployment: %w", err)
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

func (l *CKADCanaryDeployLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployments", "-l", "track=canary",
		"-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return fmt.Errorf("failed to get canary deployment: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no canary deployment found")
	}

	replicas, err := kubectl(ctx, kubeconfigPath, "get", "deployments", "-l", "track=canary",
		"-o", "jsonpath={.items[0].spec.replicas}")
	if err != nil {
		return fmt.Errorf("failed to get canary replicas: %w", err)
	}
	if strings.TrimSpace(replicas) != "1" {
		return fmt.Errorf("canary should have 1 replica, has %s", replicas)
	}
	return nil
}

func (l *CKADCanaryDeployLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create canary deployment", Command: "Create deployment webapp-canary with 1 replica and track=canary label"},
		{Description: "Verify canary exists", Command: "kubectl get deployments -l track=canary"},
		{Description: "Check service selects both", Command: "kubectl get endpoints webapp"},
	}
}
