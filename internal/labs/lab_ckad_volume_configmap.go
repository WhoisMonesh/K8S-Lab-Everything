package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADVolumeConfigMapLab{})
}

type CKADVolumeConfigMapLab struct {
	BaseLab
}

func (l *CKADVolumeConfigMapLab) ID() string             { return "ckad_volume_configmap" }
func (l *CKADVolumeConfigMapLab) Title() string          { return "Mount ConfigMap as Volume" }
func (l *CKADVolumeConfigMapLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADVolumeConfigMapLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADVolumeConfigMapLab) Cert() Cert             { return CertCKAD }
func (l *CKADVolumeConfigMapLab) DomainWeight() int      { return 20 }
func (l *CKADVolumeConfigMapLab) EstimatedTime() int     { return 15 }
func (l *CKADVolumeConfigMapLab) Tags() []string {
	return []string{"configmap", "volumes", "configuration"}
}

func (l *CKADVolumeConfigMapLab) Description() string {
	return `An application needs to read configuration from files mounted as a volume.
A ConfigMap exists but is not properly mounted to the pod.

Your task: Mount the ConfigMap 'app-config' as a volume in the pod.`
}

func (l *CKADVolumeConfigMapLab) Hints() []string {
	return []string{
		"Use volumes with configMap type",
		"Mount the volume at the path the application expects",
		"Check the ConfigMap keys to understand the file names",
	}
}

func (l *CKADVolumeConfigMapLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADVolumeConfigMapLab) Break(ctx context.Context, kubeconfigPath string) error {
	cm := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  config.yaml: |
    database:
      host: localhost
      port: 5432
  log-level: info`
	if err := kubectlApply(ctx, kubeconfigPath, cm); err != nil {
		return fmt.Errorf("creating configmap: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: app-server
  labels:
    app: app-server
spec:
  containers:
  - name: app
    image: nginx:alpine
    volumeMounts: []`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADVolumeConfigMapLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "app-server",
		"-o", "jsonpath={.spec.volumes[*].configMap.name}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no ConfigMap volume found")
	}
	if !strings.Contains(output, "app-config") {
		return fmt.Errorf("ConfigMap 'app-config' not mounted")
	}
	return nil
}

func (l *CKADVolumeConfigMapLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check ConfigMap", Command: "kubectl get configmap app-config -o yaml"},
		{Description: "Edit pod to mount ConfigMap", Command: "kubectl edit pod app-server"},
		{Description: "Add volume and volumeMount", Command: "Add configMap volume and mount at /app/config"},
		{Description: "Verify mount", Command: "kubectl exec app-server -- ls /app/config"},
	}
}
