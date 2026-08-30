package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&ImagePullBackoffCKALab{})
}

type ImagePullBackoffCKALab struct {
	BaseLab
}

func (l *ImagePullBackoffCKALab) ID() string { return "cka_image_pull_backoff" }
func (l *ImagePullBackoffCKALab) Title() string {
	return "Debug ImagePullBackOff Errors"
}
func (l *ImagePullBackoffCKALab) Category() Category     { return CategoryTroubleshooting }
func (l *ImagePullBackoffCKALab) Difficulty() Difficulty { return DifficultyEasy }
func (l *ImagePullBackoffCKALab) EstimatedTime() int     { return 15 }
func (l *ImagePullBackoffCKALab) Tags() []string {
	return []string{"image", "pull", "backoff", "troubleshooting"}
}
func (l *ImagePullBackoffCKALab) Cert() Cert        { return CertCKA }
func (l *ImagePullBackoffCKALab) DomainWeight() int { return 30 }

func (l *ImagePullBackoffCKALab) Description() string {
	return `A pod is in ImagePullBackOff state. Determine why the image cannot
be pulled and fix the image reference or create an image pull secret.`
}

func (l *ImagePullBackoffCKALab) Hints() []string {
	return []string{
		"Check pod events for pull errors",
		"Verify the image name and tag",
		"Check if imagePullSecrets are needed",
	}
}

func (l *ImagePullBackoffCKALab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ImagePullBackoffCKALab) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: image-ns
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: image-app
  namespace: image-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: web
        image: nginx:9.9.9
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, manifest); err != nil {
		return fmt.Errorf("creating broken deployment: %w", err)
	}
	return nil
}

func (l *ImagePullBackoffCKALab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ImagePullBackoffCKALab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "image-ns",
		"-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("pod not in Running state")
	}
	return nil
}

func (l *ImagePullBackoffCKALab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod status", Command: "kubectl get pods -n image-ns"},
		{Description: "Check events", Command: "kubectl describe pod -n image-ns <pod-name>"},
		{Description: "Fix image name", Command: "Correct the image reference or create pull secret"},
		{Description: "Verify", Command: "kubectl get pods -n image-ns"},
	}
}
