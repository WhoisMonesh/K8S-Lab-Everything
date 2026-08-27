package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ConfigMapBinaryDataWrongLab{})
}

type ConfigMapBinaryDataWrongLab struct {
	BaseLab
}

func (l *ConfigMapBinaryDataWrongLab) ID() string {
	return "configmap_binary_data_wrong"
}

func (l *ConfigMapBinaryDataWrongLab) Title() string {
	return "ConfigMap binaryData Corruption"
}

func (l *ConfigMapBinaryDataWrongLab) Category() Category {
	return CategoryWorkloads
}

func (l *ConfigMapBinaryDataWrongLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ConfigMapBinaryDataWrongLab) Description() string {
	return `A ConfigMap 'app-binary' has corrupted binaryData. The base64-encoded
data is invalid and causes pods using it to fail when trying to mount
the volume.

Your task: Fix the ConfigMap binaryData with valid base64 encoding.`
}

func (l *ConfigMapBinaryDataWrongLab) Hints() []string {
	return []string{
		"Check the ConfigMap data",
		"binaryData must contain valid base64-encoded values",
		"Use 'echo -n \"content\" | base64' to generate valid base64",
	}
}

func (l *ConfigMapBinaryDataWrongLab) EstimatedTime() int {
	return 15
}

func (l *ConfigMapBinaryDataWrongLab) Tags() []string {
	return []string{"configmap", "binary-data", "base64", "troubleshooting"}
}

func (l *ConfigMapBinaryDataWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ConfigMapBinaryDataWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	configMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-binary
  namespace: default
binaryData:
  config.bin: not-valid-base64!!@#$%
`
	if err := kubectlApply(ctx, kubeconfigPath, configMap); err != nil {
		return fmt.Errorf("creating ConfigMap: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: binary-app
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'cat /etc/config/config.bin && sleep 3600']
    volumeMounts:
    - name: config-vol
      mountPath: /etc/config
  volumes:
  - name: config-vol
    configMap:
      name: app-binary
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *ConfigMapBinaryDataWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ConfigMapBinaryDataWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "binary-app",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *ConfigMapBinaryDataWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod binary-app",
			Notes:       "Pod should be in CreateError or CrashLoopBackOff",
		},
		{
			Description: "Check ConfigMap",
			Command:     "kubectl get configmap app-binary -o yaml",
			Notes:       "binaryData contains invalid base64",
		},
		{
			Description: "Fix the binaryData",
			Command:     "kubectl edit configmap app-binary",
			Notes:       "Replace invalid base64 with valid encoded data",
		},
		{
			Description: "Delete and recreate pod",
			Command:     "kubectl delete pod binary-app && kubectl apply -f - <<EOF\napiVersion: v1\nkind: Pod\nmetadata:\n  name: binary-app\n  namespace: default\nspec:\n  containers:\n  - name: app\n    image: busybox:1.36\n    command: ['sh', '-c', 'cat /etc/config/config.bin && sleep 3600']\n    volumeMounts:\n    - name: config-vol\n      mountPath: /etc/config\n  volumes:\n  - name: config-vol\n    configMap:\n      name: app-binary\nEOF",
			Notes:       "Pod should now mount the config successfully",
		},
	}
}
