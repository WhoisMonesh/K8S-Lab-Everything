package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodSelectorNoMatch{})
}

type PodSelectorNoMatch struct {
	BaseLab
}

func (l *PodSelectorNoMatch) ID() string             { return "pod_selector_no_match" }
func (l *PodSelectorNoMatch) Title() string          { return "Deployment Selector Doesn't Match Labels" }
func (l *PodSelectorNoMatch) Category() Category     { return CategoryWorkloads }
func (l *PodSelectorNoMatch) Difficulty() Difficulty { return DifficultyMedium }
func (l *PodSelectorNoMatch) EstimatedTime() int     { return 15 }
func (l *PodSelectorNoMatch) Tags() []string         { return []string{"deployment", "selector", "labels"} }

func (l *PodSelectorNoMatch) Description() string {
	return `A deployment cannot manage its pods because the selector doesn't match the pod labels.
Fix the selector or pod labels to match.`
}

func (l *PodSelectorNoMatch) Hints() []string {
	return []string{
		"Check the deployment selector",
		"Compare with pod template labels",
		"Ensure selector matches labels exactly",
	}
}

func (l *PodSelectorNoMatch) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodSelectorNoMatch) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: mismatch-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: mismatch-app
      env: production
  template:
    metadata:
      labels:
        app: mismatch-app
        env: staging
    spec:
      containers:
      - name: nginx
        image: nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PodSelectorNoMatch) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/mismatch-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return err
	}
	if output != "2" {
		return fmt.Errorf("deployment not ready")
	}
	return nil
}

func (l *PodSelectorNoMatch) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check selector", Command: "kubectl get deploy mismatch-app -o jsonpath='{.spec.selector}'"},
		{Description: "Check pod labels", Command: "kubectl get pods --show-labels"},
		{Description: "Fix labels", Command: "kubectl patch deploy mismatch-app -p '{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"env\":\"production\"}}}}}'"},
	}
}
