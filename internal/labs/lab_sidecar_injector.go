package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&SidecarInjector{})
}

type SidecarInjector struct {
	BaseLab
}

func (l *SidecarInjector) ID() string            { return "sidecar_injector" }
func (l *SidecarInjector) Title() string         { return "Sidecar Injection Not Working" }
func (l *SidecarInjector) Category() Category    { return CategoryWorkloads }
func (l *SidecarInjector) Difficulty() Difficulty { return DifficultyMedium }
func (l *SidecarInjector) EstimatedTime() int    { return 20 }
func (l *SidecarInjector) Tags() []string        { return []string{"sidecar", "injection", "istio"} }

func (l *SidecarInjector) Description() string {
	return `A pod should have a sidecar container injected by Istio but it's not being injected.
Debug and fix the sidecar injection issue.`
}

func (l *SidecarInjector) Hints() []string {
	return []string{
		"Check the namespace labels",
		"Verify istio-injection label is set",
		"Check pod annotations",
	}
}

func (l *SidecarInjector) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SidecarInjector) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: istio-app
  labels:
    istio-injection: disabled
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
  namespace: istio-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *SidecarInjector) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "istio-app",
		"-o", "jsonpath={.items[0].spec.containers[*].name}")
	if err != nil {
		return err
	}
	if !containsAny(output, "istio-proxy") {
		return fmt.Errorf("sidecar not injected")
	}
	return nil
}

func (l *SidecarInjector) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check namespace labels", Command: "kubectl get namespace istio-app --show-labels"},
		{Description: "Enable injection", Command: "kubectl label namespace istio-app istio-injection=enabled --overwrite"},
		{Description: "Restart pods", Command: "kubectl rollout restart deployment webapp -n istio-app"},
	}
}
