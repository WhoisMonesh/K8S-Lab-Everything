package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ImagePullBackoff2{})
}

type ImagePullBackoff2 struct {
	BaseLab
}

func (l *ImagePullBackoff2) ID() string             { return "image_pull_backoff2" }
func (l *ImagePullBackoff2) Title() string          { return "ImagePullBackOff Due to Wrong Registry" }
func (l *ImagePullBackoff2) Category() Category     { return CategoryWorkloads }
func (l *ImagePullBackoff2) Difficulty() Difficulty { return DifficultyEasy }
func (l *ImagePullBackoff2) EstimatedTime() int     { return 10 }
func (l *ImagePullBackoff2) Tags() []string         { return []string{"images", "registry", "pull"} }

func (l *ImagePullBackoff2) Description() string {
	return `A pod is stuck in ImagePullBackOff because the image is being pulled from the wrong registry.
Fix the image reference to use the correct registry.`
}

func (l *ImagePullBackoff2) Hints() []string {
	return []string{
		"Check the pod image",
		"Look for the registry prefix",
		"Fix the image to use the correct registry",
	}
}

func (l *ImagePullBackoff2) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ImagePullBackoff2) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: wrong-registry
spec:
  replicas: 1
  selector:
    matchLabels:
      app: wrong-registry
  template:
    metadata:
      labels:
        app: wrong-registry
    spec:
      containers:
      - name: nginx
        image: my-private-registry.com/nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ImagePullBackoff2) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/wrong-registry",
		"-o", "jsonpath={.spec.template.spec.containers[0].image}")
	if err != nil {
		return err
	}
	if containsAny(output, "my-private-registry.com") {
		return fmt.Errorf("image still from wrong registry")
	}
	return nil
}

func (l *ImagePullBackoff2) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check image", Command: "kubectl get deploy wrong-registry -o jsonpath='{.spec.template.spec.containers[0].image}'"},
		{Description: "Fix image", Command: "kubectl set image deploy wrong-registry nginx=nginx:alpine"},
	}
}
