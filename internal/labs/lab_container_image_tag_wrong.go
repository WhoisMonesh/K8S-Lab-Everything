package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ContainerImageTagWrongLab{})
}

type ContainerImageTagWrongLab struct {
	BaseLab
}

func (l *ContainerImageTagWrongLab) ID() string {
	return "container_image_tag_wrong"
}

func (l *ContainerImageTagWrongLab) Title() string {
	return "Container Image Tag Wrong"
}

func (l *ContainerImageTagWrongLab) Category() Category {
	return CategoryWorkloads
}

func (l *ContainerImageTagWrongLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *ContainerImageTagWrongLab) Description() string {
	return `A deployment 'frontend' has pods stuck in ImagePullBackOff.
The image tag used doesn't exist in the registry.

Your task: Fix the deployment to use a valid image tag.`
}

func (l *ContainerImageTagWrongLab) Hints() []string {
	return []string{
		"Check the pod status",
		"Look at the image being used",
		"The image tag might be a typo or non-existent version",
		"Fix the image tag to a valid version",
	}
}

func (l *ContainerImageTagWrongLab) EstimatedTime() int {
	return 10
}

func (l *ContainerImageTagWrongLab) Tags() []string {
	return []string{"image", "deployment", "imagepullbackoff", "workloads"}
}

func (l *ContainerImageTagWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ContainerImageTagWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create deployment with non-existent image tag
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: web
        image: nginx:1.99.99-doesnotexist
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	return nil
}

func (l *ContainerImageTagWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *ContainerImageTagWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if deployment has all replicas ready
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "frontend",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	if strings.TrimSpace(output) != "2" {
		return fmt.Errorf("deployment not fully ready (ready replicas: %s, expected: 2)", output)
	}

	// Check that pods are running
	output, err = kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=frontend",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pods: %w", err)
	}

	for _, phase := range strings.Fields(output) {
		if phase != "Running" {
			return fmt.Errorf("not all pods are running (phases: %s)", output)
		}
	}

	return nil
}

func (l *ContainerImageTagWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check deployment status",
			Command:     "kubectl get deployment frontend",
			Notes:       "Replicas won't be available",
		},
		{
			Description: "Check pod status",
			Command:     "kubectl get pods -l app=frontend",
			Notes:       "Pods will be in ImagePullBackOff state",
		},
		{
			Description: "Describe a pod for details",
			Command:     "kubectl describe pod -l app=frontend | grep -A 5 Events",
			Notes:       "Look for 'Failed to pull image' error",
		},
		{
			Description: "Fix the image tag",
			Command:     "kubectl set image deployment frontend web=nginx:1.26-alpine",
			Notes:       "Update to a valid nginx image tag",
		},
		{
			Description: "Wait for rollout to complete",
			Command:     "kubectl rollout status deployment frontend",
			Notes:       "Wait for the new pods to start successfully",
		},
		{
			Description: "Verify pods are running",
			Command:     "kubectl get pods -l app=frontend",
			Notes:       "All pods should now be in Running state",
		},
	}
}
