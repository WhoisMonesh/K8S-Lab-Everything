package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&HPANotScalingLab{})
}

type HPANotScalingLab struct {
	BaseLab
}

func (l *HPANotScalingLab) ID() string {
	return "hpa_not_scaling"
}

func (l *HPANotScalingLab) Title() string {
	return "Horizontal Pod Autoscaler Not Scaling"
}

func (l *HPANotScalingLab) Category() Category {
	return CategoryAppDeployment
}

func (l *HPANotScalingLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *HPANotScalingLab) Description() string {
	return `A Horizontal Pod Autoscaler has been configured but is not scaling
the deployment despite increased load. The HPA shows current replicas
equal to desired and never scales up.

Your task: Diagnose why the HPA is not scaling and fix the issue.`
}

func (l *HPANotScalingLab) Hints() []string {
	return []string{
		"Check if metrics-server is running",
		"Verify the HPA configuration and targets",
		"Check if the deployment has resource requests defined",
	}
}

func (l *HPANotScalingLab) EstimatedTime() int {
	return 25
}

func (l *HPANotScalingLab) Tags() []string {
	return []string{"hpa", "autoscaling", "metrics", "workloads"}
}

func (l *HPANotScalingLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *HPANotScalingLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: scale-target
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: scale-target
  template:
    metadata:
      labels:
        app: scale-target
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo scaling; sleep 15; done']
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return err
	}

	hpa := `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: scale-target-hpa
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: scale-target
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
	return kubectlApply(ctx, kubeconfigPath, hpa)
}

func (l *HPANotScalingLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "hpa", "scale-target-hpa",
		"-o", "jsonpath={.status.currentReplicas}")
	if err != nil {
		return fmt.Errorf("checking hpa: %w", err)
	}

	current := strings.TrimSpace(output)
	if current == "1" || current == "" {
		return nil
	}

	return fmt.Errorf("HPA is scaling (expected stuck)")
}

func (l *HPANotScalingLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "scale-target",
		"-o", "jsonpath={.spec.template.spec.containers[0].resources}")
	if err != nil {
		return fmt.Errorf("checking deployment: %w", err)
	}

	if !strings.Contains(output, "requests") {
		return fmt.Errorf("deployment still missing resource requests")
	}

	return nil
}

func (l *HPANotScalingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check HPA status",
			Command:     "kubectl get hpa scale-target-hpa",
			Notes:       "Look at TARGETS column - shows <unknown> or 0%",
		},
		{
			Description: "Check if metrics-server is running",
			Command:     "kubectl get pods -n kube-system -l k8s-app=metrics-server",
			Notes:       "Metrics-server is required for HPA",
		},
		{
			Description: "Fix: Add resource requests to deployment",
			Command:     `kubectl patch deploy scale-target --type='json' -p='[{"op":"add","path":"/spec/template/spec/containers/0/resources","value":{"requests":{"cpu":"100m","memory":"64Mi"},"limits":{"cpu":"200m","memory":"128Mi"}}}]'`,
			Notes:       "HPA needs resource requests to calculate utilization",
		},
		{
			Description: "Verify HPA is now active",
			Command:     "kubectl get hpa scale-target-hpa",
			Notes:       "TARGETS should show a percentage instead of <unknown>",
		},
	}
}
