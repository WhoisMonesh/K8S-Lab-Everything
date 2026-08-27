package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&LimitRangeDefaultContainerLab{})
}

type LimitRangeDefaultContainerLab struct {
	BaseLab
}

func (l *LimitRangeDefaultContainerLab) ID() string {
	return "limitrange_default_container"
}

func (l *LimitRangeDefaultContainerLab) Title() string {
	return "LimitRange Default Container Limits"
}

func (l *LimitRangeDefaultContainerLab) Category() Category {
	return CategoryScheduling
}

func (l *LimitRangeDefaultContainerLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *LimitRangeDefaultContainerLab) Description() string {
	return `A LimitRange 'default-limits' has default limits that are too low
(50m CPU, 64Mi memory). Pods are being OOMKilled or throttled because
the default limits don't match application requirements.

Your task: Update the LimitRange defaults to match application needs.`
}

func (l *LimitRangeDefaultContainerLab) Hints() []string {
	return []string{
		"Check the LimitRange configuration",
		"Default limits are applied to containers without explicit limits",
		"Increase CPU and memory limits",
	}
}

func (l *LimitRangeDefaultContainerLab) EstimatedTime() int {
	return 15
}

func (l *LimitRangeDefaultContainerLab) Tags() []string {
	return []string{"limitrange", "limits", "scheduling"}
}

func (l *LimitRangeDefaultContainerLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *LimitRangeDefaultContainerLab) Break(ctx context.Context, kubeconfigPath string) error {
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: limited-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	limitRange := `apiVersion: v1
kind: LimitRange
metadata:
  name: default-limits
  namespace: limited-ns
spec:
  limits:
  - type: Container
    default:
      cpu: 50m
      memory: 64Mi
    defaultRequest:
      cpu: 25m
      memory: 32Mi
`
	if err := kubectlApply(ctx, kubeconfigPath, limitRange); err != nil {
		return fmt.Errorf("creating limitrange: %w", err)
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: memory-hungry
  namespace: limited-ns
spec:
  replicas: 2
  selector:
    matchLabels:
      app: memory-hungry
  template:
    metadata:
      labels:
        app: memory-hungry
    spec:
      containers:
      - name: app
        image: nginx:alpine
        resources: {}
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	return nil
}

func (l *LimitRangeDefaultContainerLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "limited-ns",
		"-l", "app=memory-hungry", "-o", "jsonpath={.items[*].status.containerStatuses[*].lastState.terminated.reason}")
	if strings.Contains(output, "OOMKilled") {
		return nil
	}
	return nil
}

func (l *LimitRangeDefaultContainerLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "limitrange", "default-limits",
		"-n", "limited-ns", "-o", "jsonpath={.spec.limits[0].default.cpu}")
	if err != nil {
		return fmt.Errorf("failed to check limitrange: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "50m" {
		return fmt.Errorf("CPU limit is still 50m")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "limited-ns",
		"-l", "app=memory-hungry", "-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pods: %w", err)
	}

	for _, phase := range splitFields(output) {
		if phase != "Running" {
			return fmt.Errorf("not all pods are running")
		}
	}

	return nil
}

func (l *LimitRangeDefaultContainerLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check LimitRange",
			Command:     "kubectl get limitrange default-limits -n limited-ns -o yaml",
			Notes:       "Default CPU is 50m, memory is 64Mi - too low",
		},
		{
			Description: "Check pod status",
			Command:     "kubectl get pods -n limited-ns",
			Notes:       "Pods may be OOMKilled or throttled",
		},
		{
			Description: "Fix LimitRange defaults",
			Command:     "kubectl edit limitrange default-limits -n limited-ns",
			Notes:       "Increase default CPU to 200m and memory to 256Mi",
		},
		{
			Description: "Verify pods are running",
			Command:     "kubectl get pods -n limited-ns",
			Notes:       "All pods should be Running",
		},
	}
}
