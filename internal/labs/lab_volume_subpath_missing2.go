package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&VolumeSubpathMissing{})
}

type VolumeSubpathMissing struct {
	BaseLab
}

func (l *VolumeSubpathMissing) ID() string            { return "volume_subpath_missing2" }
func (l *VolumeSubpathMissing) Title() string         { return "Pod CrashLoop - Wrong Volume subPath" }
func (l *VolumeSubpathMissing) Category() Category    { return CategoryStorage }
func (l *VolumeSubpathMissing) Difficulty() Difficulty { return DifficultyMedium }
func (l *VolumeSubpathMissing) EstimatedTime() int    { return 15 }
func (l *VolumeSubpathMissing) Tags() []string        { return []string{"storage", "subpath", "volumes"} }

func (l *VolumeSubpathMissing) Description() string {
	return `A pod is crashing because the volume subPath doesn't exist in the ConfigMap.
Fix the subPath configuration.`
}

func (l *VolumeSubpathMissing) Hints() []string {
	return []string{
		"Check the volume subPath",
		"Verify the key exists in the ConfigMap",
		"Fix the subPath to match a valid key",
	}
}

func (l *VolumeSubpathMissing) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *VolumeSubpathMissing) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  config.yaml: |
    version: "1.0"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: subpath-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: subpath-app
  template:
    metadata:
      labels:
        app: subpath-app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ["sh", "-c", "cat /etc/config/nonexistent && sleep 3600"]
        volumeMounts:
        - name: config
          mountPath: /etc/config
          subPath: nonexistent
      volumes:
      - name: config
        configMap:
          name: app-config`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *VolumeSubpathMissing) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/subpath-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return err
	}
	if output != "1" {
		return fmt.Errorf("deployment not ready")
	}
	return nil
}

func (l *VolumeSubpathMissing) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check ConfigMap keys", Command: "kubectl get configmap app-config -o yaml"},
		{Description: "Fix subPath", Command: "kubectl edit deploy subpath-app"},
		{Description: "Change subPath", Command: "Change subPath from nonexistent to config.yaml"},
	}
}
