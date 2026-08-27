package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&HPACPUTargetWrongLab{})
}

type HPACPUTargetWrongLab struct {
	BaseLab
}

func (l *HPACPUTargetWrongLab) ID() string {
	return "hpa_cpu_target_wrong"
}

func (l *HPACPUTargetWrongLab) Title() string {
	return "HPA CPU Target Misconfigured"
}

func (l *HPACPUTargetWrongLab) Category() Category {
	return CategoryWorkloads
}

func (l *HPACPUTargetWrongLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *HPACPUTargetWrongLab) Description() string {
	return `A HorizontalPodAutoscaler 'web-hpa' has an impossibly high CPU target
of 500% (averageUtilization: 500). This means the HPA will never trigger
scaling because CPU usage can't exceed 100% per pod in meaningful terms.

Your task: Fix the HPA CPU target to a realistic value.`
}

func (l *HPACPUTargetWrongLab) Hints() []string {
	return []string{
		"Check the HPA metrics configuration",
		"averageUtilization should be between 1-100",
		"A typical target is 50-80% for most workloads",
	}
}

func (l *HPACPUTargetWrongLab) EstimatedTime() int {
	return 10
}

func (l *HPACPUTargetWrongLab) Tags() []string {
	return []string{"hpa", "autoscaling", "cpu", "metrics", "workloads"}
}

func (l *HPACPUTargetWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *HPACPUTargetWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
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

	hpa := `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: web-hpa
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: web-hpa
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 500
`
	if err := kubectlApply(ctx, kubeconfigPath, hpa); err != nil {
		return fmt.Errorf("creating HPA: %w", err)
	}

	return nil
}

func (l *HPACPUTargetWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *HPACPUTargetWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "hpa", "web-hpa",
		"-o", "jsonpath={.spec.metrics[0].resource.target.averageUtilization}")
	if err != nil {
		return fmt.Errorf("failed to check HPA: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "500" {
		return fmt.Errorf("CPU target is still 500%%")
	}

	return nil
}

func (l *HPACPUTargetWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check HPA configuration",
			Command:     "kubectl get hpa web-hpa",
			Notes:       "TARGETS column shows an impossibly high value",
		},
		{
			Description: "Check HPA metrics",
			Command:     "kubectl get hpa web-hpa -o yaml | grep averageUtilization",
			Notes:       "averageUtilization is 500 which is invalid",
		},
		{
			Description: "Fix the CPU target",
			Command:     "kubectl patch hpa web-hpa --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/metrics/0/resource/target/averageUtilization\",\"value\":70}]'",
			Notes:       "Set to 70% which is a typical target",
		},
		{
			Description: "Verify HPA configuration",
			Command:     "kubectl get hpa web-hpa",
			Notes:       "TARGETS should now show a realistic percentage",
		},
	}
}
