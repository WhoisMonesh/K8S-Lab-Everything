package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&HPANotWorkingLab{})
}

type HPANotWorkingLab struct {
	BaseLab
}

func (l *HPANotWorkingLab) ID() string {
	return "hpa_not_working"
}

func (l *HPANotWorkingLab) Title() string {
	return "Horizontal Pod Autoscaler Not Scaling"
}

func (l *HPANotWorkingLab) Category() Category {
	return CategoryWorkloads
}

func (l *HPANotWorkingLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *HPANotWorkingLab) Description() string {
	return `A Horizontal Pod Autoscaler (HPA) named 'web-hpa' is configured but not scaling the deployment.
The HPA shows current metrics at 0 or unknown, and the replica count never changes.

Your task: Fix the HPA so it can properly scale the deployment based on CPU usage.`
}

func (l *HPANotWorkingLab) Hints() []string {
	return []string{
		"Check the HPA status and events",
		"Verify the HPA target reference matches the deployment",
		"Check if metrics-server is running",
		"Look at the HPA manifest for incorrect selector labels",
	}
}

func (l *HPANotWorkingLab) EstimatedTime() int {
	return 20
}

func (l *HPANotWorkingLab) Tags() []string {
	return []string{"hpa", "autoscaling", "metrics", "workloads"}
}

func (l *HPANotWorkingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *HPANotWorkingLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create deployment
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-hpa
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web-hpa
  template:
    metadata:
      labels:
        app: web-hpa
    spec:
      containers:
      - name: web
        image: nginx:alpine
        resources:
          requests:
            cpu: 100m
          limits:
            cpu: 200m
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	// Create HPA with wrong target selector (selector doesn't match deployment)
	hpa := `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: web-hpa
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: wrong-deployment-name
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 50
`
	if err := kubectlApply(ctx, kubeconfigPath, hpa); err != nil {
		return fmt.Errorf("creating HPA: %w", err)
	}

	return nil
}

func (l *HPANotWorkingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *HPANotWorkingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if HPA targets the correct deployment
	output, err := kubectl(ctx, kubeconfigPath, "get", "hpa", "web-hpa",
		"-o", "jsonpath={.spec.scaleTargetRef.name}")
	if err != nil {
		return fmt.Errorf("failed to check HPA: %w", err)
	}

	if strings.TrimSpace(output) != "web-hpa" {
		return fmt.Errorf("HPA target is wrong (got: %s, expected: web-hpa)", output)
	}

	// Check if HPA has valid targets
	output, err = kubectl(ctx, kubeconfigPath, "get", "hpa", "web-hpa",
		"-o", "jsonpath={.status.currentMetrics}")
	if err != nil {
		return fmt.Errorf("failed to check HPA metrics: %w", err)
	}

	if strings.Contains(output, "unknown") || output == "[]" || output == "" {
		return fmt.Errorf("HPA metrics are still unknown or empty")
	}

	return nil
}

func (l *HPANotWorkingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check HPA status",
			Command:     "kubectl get hpa web-hpa",
			Notes:       "Notice the TARGETS column shows '<unknown>' or 0%",
		},
		{
			Description: "Describe HPA for details",
			Command:     "kubectl describe hpa web-hpa",
			Notes:       "Look for events showing unable to find target deployment",
		},
		{
			Description: "Check HPA spec",
			Command:     "kubectl get hpa web-hpa -o yaml",
			Notes:       "The scaleTargetRef.name is set to 'wrong-deployment-name' instead of 'web-hpa'",
		},
		{
			Description: "Fix the HPA target reference",
			Command:     "kubectl edit hpa web-hpa",
			Notes:       "Change spec.scaleTargetRef.name from 'wrong-deployment-name' to 'web-hpa'",
		},
		{
			Description: "Verify HPA is now targeting correctly",
			Command:     "kubectl get hpa web-hpa",
			Notes:       "TARGETS should now show a valid percentage like 0% or a number",
		},
	}
}
