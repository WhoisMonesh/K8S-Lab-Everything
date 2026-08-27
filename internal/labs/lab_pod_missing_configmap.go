package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodMissingConfigMapLab{})
}

type PodMissingConfigMapLab struct {
	BaseLab
}

func (l *PodMissingConfigMapLab) ID() string {
	return "pod_missing_configmap"
}

func (l *PodMissingConfigMapLab) Title() string {
	return "Pod Failing Due to Missing ConfigMap Mount"
}

func (l *PodMissingConfigMapLab) Category() Category {
	return CategoryWorkloads
}

func (l *PodMissingConfigMapLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *PodMissingConfigMapLab) Description() string {
	return `A pod 'config-app' is failing because it tries to mount a ConfigMap that doesn't exist.

Your task: Create the missing ConfigMap or fix the pod to not require it.`
}

func (l *PodMissingConfigMapLab) Hints() []string {
	return []string{
		"Check the pod status and events",
		"Look at the volume mounts in the pod spec",
		"Check if the referenced ConfigMap exists",
		"Create the ConfigMap or fix the volume reference",
	}
}

func (l *PodMissingConfigMapLab) EstimatedTime() int {
	return 10
}

func (l *PodMissingConfigMapLab) Tags() []string {
	return []string{"configmap", "volume", "pod", "troubleshooting"}
}

func (l *PodMissingConfigMapLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodMissingConfigMapLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: config-app
  namespace: default
spec:
  containers:
  - name: app
    image: nginx:alpine
    volumeMounts:
    - name: app-config
      mountPath: /etc/config
  volumes:
  - name: app-config
    configMap:
      name: nonexistent-config
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}
	return nil
}

func (l *PodMissingConfigMapLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodMissingConfigMapLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "config-app",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}
	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}
	return nil
}

func (l *PodMissingConfigMapLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod config-app",
			Notes:       "The pod is in CreateError",
		},
		{
			Description: "Check pod events",
			Command:     "kubectl describe pod config-app | grep -A 10 Events",
			Notes:       "Look for ConfigMap not found error",
		},
		{
			Description: "Create the missing ConfigMap",
			Command:     `kubectl create configmap nonexistent-config --from-literal=app.conf="server { listen 80; }"`,
			Notes:       "Create the ConfigMap the pod expects",
		},
		{
			Description: "Delete and recreate the pod",
			Command:     "kubectl delete pod config-app && kubectl apply -f pod.yaml",
			Notes:       "Restart the pod to mount the new ConfigMap",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod config-app",
			Notes:       "The pod should now be Running",
		},
	}
}
