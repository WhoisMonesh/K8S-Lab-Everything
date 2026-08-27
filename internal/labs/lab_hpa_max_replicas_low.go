package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&HPAMaxReplicasLowLab{})
}

type HPAMaxReplicasLowLab struct {
	BaseLab
}

func (l *HPAMaxReplicasLowLab) ID() string {
	return "hpa_max_replicas_low"
}

func (l *HPAMaxReplicasLowLab) Title() string {
	return "HPA maxReplicas Too Low"
}

func (l *HPAMaxReplicasLowLab) Category() Category {
	return CategoryWorkloads
}

func (l *HPAMaxReplicasLowLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *HPAMaxReplicasLowLab) Description() string {
	return `A HorizontalPodAutoscaler 'scaling-app' is configured with maxReplicas=2
but the deployment needs to scale up to handle increased load. The HPA
cannot scale beyond 2 replicas even when CPU usage is very high.

Your task: Increase the HPA maxReplicas to allow proper scaling.`
}

func (l *HPAMaxReplicasLowLab) Hints() []string {
	return []string{
		"Check the HPA configuration",
		"maxReplicas limits the maximum number of pods",
		"Set maxReplicas to a higher value like 10",
	}
}

func (l *HPAMaxReplicasLowLab) EstimatedTime() int {
	return 10
}

func (l *HPAMaxReplicasLowLab) Tags() []string {
	return []string{"hpa", "autoscaling", "maxReplicas", "workloads"}
}

func (l *HPAMaxReplicasLowLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *HPAMaxReplicasLowLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: scaling-app
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: scaling-app
  template:
    metadata:
      labels:
        app: scaling-app
    spec:
      containers:
      - name: app
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
  name: scaling-app
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: scaling-app
  minReplicas: 1
  maxReplicas: 2
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

func (l *HPAMaxReplicasLowLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *HPAMaxReplicasLowLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "hpa", "scaling-app",
		"-o", "jsonpath={.spec.maxReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check HPA: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "2" || val == "" {
		return fmt.Errorf("maxReplicas is still too low: %s", val)
	}

	return nil
}

func (l *HPAMaxReplicasLowLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check HPA configuration",
			Command:     "kubectl get hpa scaling-app",
			Notes:       "maxReplicas shows as 2",
		},
		{
			Description: "Check HPA details",
			Command:     "kubectl get hpa scaling-app -o yaml | grep maxReplicas",
			Notes:       "maxReplicas: 2 is too low for the workload",
		},
		{
			Description: "Increase maxReplicas",
			Command:     "kubectl patch hpa scaling-app --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/maxReplicas\",\"value\":10}]'",
			Notes:       "Set maxReplicas to 10 to allow proper scaling",
		},
		{
			Description: "Verify HPA configuration",
			Command:     "kubectl get hpa scaling-app",
			Notes:       "maxReplicas should now show 10",
		},
	}
}
