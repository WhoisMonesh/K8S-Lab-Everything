package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ImagePullBackoffNameLab{})
}

type ImagePullBackoffNameLab struct {
	BaseLab
}

func (l *ImagePullBackoffNameLab) ID() string {
	return "image_pull_backoff_name"
}

func (l *ImagePullBackoffNameLab) Title() string {
	return "ImagePullBackOff Due to Wrong Registry"
}

func (l *ImagePullBackoffNameLab) Category() Category {
	return CategoryWorkloads
}

func (l *ImagePullBackoffNameLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *ImagePullBackoffNameLab) Description() string {
	return `A deployment 'payments' is stuck in ImagePullBackOff because the image references a private registry that doesn't exist.

Your task: Fix the image reference to use a public image.`
}

func (l *ImagePullBackoffNameLab) Hints() []string {
	return []string{
		"Check the deployment status",
		"Look at the image being used",
		"The image prefix references a non-existent registry",
		"Change to a public image like nginx:alpine",
	}
}

func (l *ImagePullBackoffNameLab) EstimatedTime() int {
	return 10
}

func (l *ImagePullBackoffNameLab) Tags() []string {
	return []string{"image", "registry", "deployment", "imagepullbackoff"}
}

func (l *ImagePullBackoffNameLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ImagePullBackoffNameLab) Break(ctx context.Context, kubeconfigPath string) error {
	ds := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: payments
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: payments
  template:
    metadata:
      labels:
        app: payments
    spec:
      containers:
      - name: app
        image: my-private-registry.example.com/payments:v2.1
`
	if err := kubectlApply(ctx, kubeconfigPath, ds); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}
	return nil
}

func (l *ImagePullBackoffNameLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *ImagePullBackoffNameLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "payments",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}
	if strings.TrimSpace(output) != "1" {
		return fmt.Errorf("deployment not ready (ready: %s)", output)
	}
	return nil
}

func (l *ImagePullBackoffNameLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check deployment status",
			Command:     "kubectl get deployment payments",
			Notes:       "Available replicas will be 0",
		},
		{
			Description: "Check pod status",
			Command:     "kubectl get pods -l app=payments",
			Notes:       "Pod will be in ImagePullBackOff",
		},
		{
			Description: "Fix the image",
			Command:     "kubectl set image deployment payments app=nginx:alpine",
			Notes:       "Change to a valid public image",
		},
		{
			Description: "Verify rollout",
			Command:     "kubectl rollout status deployment payments",
			Notes:       "Wait for the new image to be pulled and pod to start",
		},
	}
}
