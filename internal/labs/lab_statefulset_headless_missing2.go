package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&StatefulsetHeadlessMissing{})
}

type StatefulsetHeadlessMissing struct {
	BaseLab
}

func (l *StatefulsetHeadlessMissing) ID() string             { return "statefulset_headless_missing2" }
func (l *StatefulsetHeadlessMissing) Title() string          { return "StatefulSet Without Headless Service" }
func (l *StatefulsetHeadlessMissing) Category() Category     { return CategoryWorkloads }
func (l *StatefulsetHeadlessMissing) Difficulty() Difficulty { return DifficultyMedium }
func (l *StatefulsetHeadlessMissing) EstimatedTime() int     { return 20 }
func (l *StatefulsetHeadlessMissing) Tags() []string {
	return []string{"statefulset", "headless", "dns"}
}

func (l *StatefulsetHeadlessMissing) Description() string {
	return `A StatefulSet cannot start because it references a headless service that doesn't exist.
Create the headless service for the StatefulSet.`
}

func (l *StatefulsetHeadlessMissing) Hints() []string {
	return []string{
		"Check the StatefulSet serviceName",
		"Verify the headless service exists",
		"Create a service with clusterIP: None",
	}
}

func (l *StatefulsetHeadlessMissing) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *StatefulsetHeadlessMissing) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  replicas: 3
  serviceName: web-headless
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
        image: nginx:alpine
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *StatefulsetHeadlessMissing) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "web-headless",
		"-o", "jsonpath={.spec.clusterIP}")
	if err != nil {
		return err
	}
	if output != "None" {
		return fmt.Errorf("service not headless")
	}
	return nil
}

func (l *StatefulsetHeadlessMissing) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check StatefulSet", Command: "kubectl get statefulset web -o yaml"},
		{Description: "Create headless service", Command: "kubectl expose statefulset web --name=web-headless --clusterIP=None"},
	}
}
