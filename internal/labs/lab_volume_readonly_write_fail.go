package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&VolumeReadOnlyWriteFailLab{}) }

type VolumeReadOnlyWriteFailLab struct{ BaseLab }

func (l *VolumeReadOnlyWriteFailLab) ID() string { return "volume_readonly_write_fail" }
func (l *VolumeReadOnlyWriteFailLab) Title() string {
	return "Pod CrashLoop — Writing to ReadOnly Volume"
}
func (l *VolumeReadOnlyWriteFailLab) Category() Category     { return CategoryStorage }
func (l *VolumeReadOnlyWriteFailLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *VolumeReadOnlyWriteFailLab) EstimatedTime() int     { return 10 }
func (l *VolumeReadOnlyWriteFailLab) Tags() []string {
	return []string{"volume", "readonly", "security", "storage"}
}
func (l *VolumeReadOnlyWriteFailLab) Description() string {
	return `A pod 'writer' crashes because it tries to write to a ConfigMap
volume, which is read-only by default in Kubernetes.

The app writes its PID to /data/pid.txt on startup, but the mount
is read-only.

Your task: Add an emptyDir volume (writable) alongside the read-only
ConfigMap volume. Update the pod to use the writable volume for
/data instead of /data/pid.txt on the ConfigMap mount.`
}
func (l *VolumeReadOnlyWriteFailLab) Hints() []string {
	return []string{
		"Check: kubectl describe pod writer — look for read-only file system errors",
		"ConfigMap volumes are always read-only",
		"Add an emptyDir volume and mount it at /data",
		"Keep the ConfigMap volume for config files, use emptyDir for writable data",
	}
}

func (l *VolumeReadOnlyWriteFailLab) Break(ctx context.Context, kp string) error {
	cm := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  setting: "enabled"
`
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: writer
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ["sh","-c","echo $$ > /data/pid.txt && echo 'started' && while true; do sleep 5; done"]
    volumeMounts:
    - name: config
      mountPath: /data
  volumes:
  - name: config
    configMap:
      name: app-config
`
	kubectlApply(ctx, kp, cm)
	return kubectlApply(ctx, kp, pod)
}

func (l *VolumeReadOnlyWriteFailLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *VolumeReadOnlyWriteFailLab) Verify(ctx context.Context, kp string) error {
	phase, _ := kubectl(ctx, kp, "get", "pod", "writer", "-o", "jsonpath={.status.phase}")
	if phase != "Running" {
		return fmt.Errorf("pod not running (phase: %s)", phase)
	}
	// Verify writable volume exists
	vols, _ := kubectl(ctx, kp, "get", "pod", "writer", "-o",
		"jsonpath={.spec.volumes[*].emptyDir}")
	if vols == "" {
		return fmt.Errorf("no emptyDir volume found — data still going to read-only ConfigMap")
	}
	return nil
}

func (l *VolumeReadOnlyWriteFailLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check crash reason", Command: "kubectl describe pod writer | tail -10", Notes: "Read-only file system error"},
		{Description: "Fix by adding writable emptyDir volume", Command: `kubectl delete pod writer && kubectl run writer --image=busybox:1.36 --restart=Never --overrides='{"spec":{"volumes":[{"name":"config","configMap":{"name":"app-config"}},{"name":"data","emptyDir":{}}],"containers":[{"name":"writer","image":"busybox:1.36","command":["sh","-c","echo $$ > /data/pid.txt && echo started && while true; do sleep 5; done"],"volumeMounts":[{"name":"config","mountPath":"/config"},{"name":"data","mountPath":"/data"}]}]}}'`, Notes: "Separates config (read-only) from data (writable)"},
		{Description: "Verify", Command: "kubectl get pod writer && kubectl exec writer -- cat /data/pid.txt", Notes: "Pod Running, PID file written successfully"},
	}
}
