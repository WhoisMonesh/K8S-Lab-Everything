package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADBlueGreenDeployLab{})
}

type CKADBlueGreenDeployLab struct {
	BaseLab
}

func (l *CKADBlueGreenDeployLab) ID() string             { return "ckad_blue_green_deploy" }
func (l *CKADBlueGreenDeployLab) Title() string          { return "Implement Blue/Green Deployment" }
func (l *CKADBlueGreenDeployLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADBlueGreenDeployLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKADBlueGreenDeployLab) Cert() Cert             { return CertCKAD }
func (l *CKADBlueGreenDeployLab) DomainWeight() int      { return 20 }
func (l *CKADBlueGreenDeployLab) EstimatedTime() int     { return 25 }
func (l *CKADBlueGreenDeployLab) Tags() []string {
	return []string{"blue-green", "deployment", "strategy"}
}

func (l *CKADBlueGreenDeployLab) Description() string {
	return `Implement a blue/green deployment pattern where two versions of the application
run simultaneously. Traffic should be switched from blue (v1) to green (v2)
by updating the Service selector.

Your task: Create the green deployment and update the Service to point to it.`
}

func (l *CKADBlueGreenDeployLab) Hints() []string {
	return []string{
		"Blue/green uses two identical deployments with different labels",
		"The Service selector determines which deployment receives traffic",
		"Update the Service selector to switch traffic",
	}
}

func (l *CKADBlueGreenDeployLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADBlueGreenDeployLab) Break(ctx context.Context, kubeconfigPath string) error {
	blue := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp-blue
  labels:
    app: webapp
    version: blue
spec:
  replicas: 3
  selector:
    matchLabels:
      app: webapp
      version: blue
  template:
    metadata:
      labels:
        app: webapp
        version: blue
    spec:
      containers:
      - name: webapp
        image: nginx:1.21
        ports:
        - containerPort: 80`
	if err := kubectlApply(ctx, kubeconfigPath, blue); err != nil {
		return fmt.Errorf("creating blue deployment: %w", err)
	}

	green := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp-green
  labels:
    app: webapp
    version: green
spec:
  replicas: 3
  selector:
    matchLabels:
      app: webapp
      version: green
  template:
    metadata:
      labels:
        app: webapp
        version: green
    spec:
      containers:
      - name: webapp
        image: nginx:1.25
        ports:
        - containerPort: 80`
	if err := kubectlApply(ctx, kubeconfigPath, green); err != nil {
		return fmt.Errorf("creating green deployment: %w", err)
	}

	svc := `apiVersion: v1
kind: Service
metadata:
  name: webapp
spec:
  selector:
    app: webapp
    version: blue
  ports:
  - port: 80
    targetPort: 80`
	return kubectlApply(ctx, kubeconfigPath, svc)
}

func (l *CKADBlueGreenDeployLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "webapp",
		"-o", "jsonpath={.spec.selector.version}")
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}
	if strings.TrimSpace(output) != "green" {
		return fmt.Errorf("service not pointing to green (current: %s)", output)
	}
	return nil
}

func (l *CKADBlueGreenDeployLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current service selector", Command: "kubectl get service webapp -o yaml | grep -A 3 selector"},
		{Description: "Update service to green", Command: "kubectl patch service webapp -p '{\"spec\":{\"selector\":{\"version\":\"green\"}}}'"},
		{Description: "Verify traffic switched", Command: "kubectl get service webapp -o yaml | grep version"},
	}
}
