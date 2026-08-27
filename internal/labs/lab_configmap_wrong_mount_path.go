package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ConfigMapWrongMountPathLab{})
}

type ConfigMapWrongMountPathLab struct {
	BaseLab
}

func (l *ConfigMapWrongMountPathLab) ID() string {
	return "configmap_wrong_mount_path"
}

func (l *ConfigMapWrongMountPathLab) Title() string {
	return "ConfigMap Mounted at Wrong Path"
}

func (l *ConfigMapWrongMountPathLab) Category() Category {
	return CategoryWorkloads
}

func (l *ConfigMapWrongMountPathLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *ConfigMapWrongMountPathLab) Description() string {
	return `A pod 'app-server' is failing to start because a ConfigMap is mounted at an incorrect path.
The application expects configuration files at /etc/app/config but the volume mount points elsewhere.

Your task: Fix the pod's volume mount path to match what the application expects.`
}

func (l *ConfigMapWrongMountPathLab) Hints() []string {
	return []string{
		"Check pod events for mount-related errors",
		"Compare the volumeMount mountPath with what the application expects",
		"The application expects files at /etc/app/config",
	}
}

func (l *ConfigMapWrongMountPathLab) EstimatedTime() int {
	return 10
}

func (l *ConfigMapWrongMountPathLab) Tags() []string {
	return []string{"configmap", "volume-mount", "troubleshooting"}
}

func (l *ConfigMapWrongMountPathLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ConfigMapWrongMountPathLab) Break(ctx context.Context, kubeconfigPath string) error {
	configMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  config.yaml: |
    server:
      port: 8080
      host: 0.0.0.0
`
	if err := kubectlApply(ctx, kubeconfigPath, configMap); err != nil {
		return fmt.Errorf("creating ConfigMap: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: app-server
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'cat /etc/app/config/config.yaml && sleep 3600']
    volumeMounts:
    - name: config-vol
      mountPath: /tmp/wrong-path
  volumes:
  - name: config-vol
    configMap:
      name: app-config
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *ConfigMapWrongMountPathLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ConfigMapWrongMountPathLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "app-server",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	output, err = kubectl(ctx, kubeconfigPath, "exec", "app-server",
		"--", "cat", "/etc/app/config/config.yaml")
	if err != nil {
		return fmt.Errorf("cannot read config at expected path: %w", err)
	}

	if !strings.Contains(output, "port: 8080") {
		return fmt.Errorf("config content not found at expected path")
	}

	return nil
}

func (l *ConfigMapWrongMountPathLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod events",
			Command:     "kubectl describe pod app-server | grep -A 10 Events",
			Notes:       "Look for mount or file not found errors",
		},
		{
			Description: "Check the volume mount configuration",
			Command:     "kubectl get pod app-server -o yaml | grep -A 5 volumeMounts",
			Notes:       "mountPath is /tmp/wrong-path instead of /etc/app/config",
		},
		{
			Description: "Fix the volume mount path",
			Command:     "kubectl edit pod app-server",
			Notes:       "Change mountPath from /tmp/wrong-path to /etc/app/config",
		},
		{
			Description: "Verify the fix",
			Command:     "kubectl get pod app-server",
			Notes:       "Pod should be Running and config accessible at /etc/app/config/config.yaml",
		},
	}
}
