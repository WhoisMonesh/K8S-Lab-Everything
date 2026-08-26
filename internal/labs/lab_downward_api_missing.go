package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&DownwardAPIMissing{})
}

type DownwardAPIMissing struct {
	BaseLab
}

func (l *DownwardAPIMissing) ID() string             { return "downward_api_missing" }
func (l *DownwardAPIMissing) Title() string          { return "DownwardAPI Volume Missing" }
func (l *DownwardAPIMissing) Category() Category     { return CategoryWorkloads }
func (l *DownwardAPIMissing) Difficulty() Difficulty { return DifficultyMedium }
func (l *DownwardAPIMissing) EstimatedTime() int     { return 15 }
func (l *DownwardAPIMissing) Tags() []string         { return []string{"downwardapi", "volumes", "workloads"} }

func (l *DownwardAPIMissing) Description() string {
	return `A pod is crashing because it expects metadata from a DownwardAPI volume that is misconfigured.
Fix the volume mount to provide the correct metadata.`
}

func (l *DownwardAPIMissing) Hints() []string {
	return []string{
		"Check the pod volumes",
		"Look at the downwardAPI field",
		"Verify the items list includes correct field paths",
	}
}

func (l *DownwardAPIMissing) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DownwardAPIMissing) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: downward-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: downward-app
  template:
    metadata:
      labels:
        app: downward-app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ["sh", "-c", "cat /etc/podinfo/labels && sleep 3600"]
        volumeMounts:
        - name: podinfo
          mountPath: /etc/podinfo
      volumes:
      - name: podinfo
        downwardAPI:
          items:
          - path: "labels"
            fieldRef:
              fieldPath: metadata.labels.nonexistent`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *DownwardAPIMissing) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/downward-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return err
	}
	if output != "1" {
		return fmt.Errorf("deployment not ready")
	}
	return nil
}

func (l *DownwardAPIMissing) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod events", Command: "kubectl describe deploy downward-app"},
		{Description: "Fix downwardAPI field path", Command: "kubectl edit deploy downward-app"},
		{Description: "Use correct field path", Command: "Change fieldPath from metadata.labels.nonexistent to metadata.labels.app"},
	}
}
