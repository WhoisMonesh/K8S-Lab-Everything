package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&ImagePullBackOffLab{})
}

type ImagePullBackOffLab struct {
	BaseLab
}

func (l *ImagePullBackOffLab) ID() string {
	return "image_pull_backoff"
}

func (l *ImagePullBackOffLab) Title() string {
	return "ImagePullBackOff Error"
}

func (l *ImagePullBackOffLab) Category() Category {
	return CategoryWorkloads
}

func (l *ImagePullBackOffLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *ImagePullBackOffLab) Description() string {
	return `A deployment named 'web-app' cannot start its pods.
All pods are in ImagePullBackOff state.

Your task: Fix the issue so the pods can pull the image and start running.`
}

func (l *ImagePullBackOffLab) Hints() []string {
	return []string{
		"Check the pod status and events",
		"Look at the image name in the pod spec",
		"The image name might have a typo",
		"Check the image tag or registry",
	}
}

func (l *ImagePullBackOffLab) EstimatedTime() int {
	return 10
}

func (l *ImagePullBackOffLab) Tags() []string {
	return []string{"pods", "images", "troubleshooting", "image-pull", "deployments"}
}

func (l *ImagePullBackOffLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ImagePullBackOffLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create a deployment with a wrong image name
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: nginx
        image: ngixn:alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	return nil
}

func (l *ImagePullBackOffLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Wait for pods to enter ImagePullBackOff
	time.Sleep(15 * time.Second)
	return nil
}

func (l *ImagePullBackOffLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if pods are running
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "web-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("checking deployment: %w", err)
	}

	if output == "2" {
		return nil
	}

	return fmt.Errorf("deployment not fully ready yet")
}

func (l *ImagePullBackOffLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check the deployment status",
			Command:     "kubectl get deployment web-app",
			Notes:       "Notice that 0/2 replicas are ready",
		},
		{
			Description: "Check the pod status",
			Command:     "kubectl get pods -l app=web",
			Notes:       "Pods should be in ImagePullBackOff or ErrImagePull state",
		},
		{
			Description: "Describe a pod to see the error",
			Command:     "kubectl describe pod -l app=web | grep -A 5 Events",
			Notes:       "Look for 'Failed to pull image' error with details about the image",
		},
		{
			Description: "Identify the typo",
			Command:     "kubectl get deployment web-app -o jsonpath='{.spec.template.spec.containers[0].image}'",
			Notes:       "The image is 'ngixn:alpine' (typo: should be 'nginx', not 'ngixn')",
		},
		{
			Description: "Fix the deployment image",
			Command:     "kubectl set image deployment/web-app nginx=nginx:alpine",
			Notes:       "This updates the container image to the correct name",
		},
		{
			Description: "Alternative: Edit the deployment directly",
			Command:     "kubectl edit deployment web-app",
			Notes:       "Change 'ngixn:alpine' to 'nginx:alpine' in the editor",
		},
		{
			Description: "Verify the fix",
			Command:     "kubectl get pods -l app=web",
			Notes:       "Pods should now be in Running state",
		},
		{
			Description: "Check rollout status",
			Command:     "kubectl rollout status deployment/web-app",
			Notes:       "Should show 'successfully rolled out'",
		},
	}
}
