package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&VolumeReadOnlyWriteFail2{})
}

type VolumeReadOnlyWriteFail2 struct {
	BaseLab
}

func (l *VolumeReadOnlyWriteFail2) ID() string            { return "volume_readonly_write_fail2" }
func (l *VolumeReadOnlyWriteFail2) Title() string         { return "Pod CrashLoop - Writing to ReadOnly Volume" }
func (l *VolumeReadOnlyWriteFail2) Category() Category    { return CategoryStorage }
func (l *VolumeReadOnlyWriteFail2) Difficulty() Difficulty { return DifficultyEasy }
func (l *VolumeReadOnlyWriteFail2) EstimatedTime() int    { return 10 }
func (l *VolumeReadOnlyWriteFail2) Tags() []string        { return []string{"storage", "readonly", "volumes"} }

func (l *VolumeReadOnlyWriteFail2) Description() string {
	return `A pod is crashing because it's trying to write to a read-only volume mount.
Fix the volume mount to allow writes.`
}

func (l *VolumeReadOnlyWriteFail2) Hints() []string {
	return []string{
		"Check the volume mount readOnly setting",
		"Look at the container volumeMounts",
		"Remove or set readOnly to false",
	}
}

func (l *VolumeReadOnlyWriteFail2) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *VolumeReadOnlyWriteFail2) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: readonly-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: readonly-app
  template:
    metadata:
      labels:
        app: readonly-app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ["sh", "-c", "echo test > /data/file.txt && sleep 3600"]
        volumeMounts:
        - name: data
          mountPath: /data
          readOnly: true
      volumes:
      - name: data
        emptyDir: {}`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *VolumeReadOnlyWriteFail2) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/readonly-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return err
	}
	if output != "1" {
		return fmt.Errorf("deployment not ready")
	}
	return nil
}

func (l *VolumeReadOnlyWriteFail2) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check volume mounts", Command: "kubectl get deploy readonly-app -o jsonpath='{.spec.template.spec.containers[0].volumeMounts}'"},
		{Description: "Fix readOnly", Command: "kubectl edit deploy readonly-app"},
		{Description: "Remove readOnly or set to false", Command: "Change readOnly: true to readOnly: false or remove the line"},
	}
}
