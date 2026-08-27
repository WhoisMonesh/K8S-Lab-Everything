package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADVolumeEmptyDirLab{})
}

type CKADVolumeEmptyDirLab struct {
	BaseLab
}

func (l *CKADVolumeEmptyDirLab) ID() string             { return "ckad_volume_emptydir" }
func (l *CKADVolumeEmptyDirLab) Title() string          { return "Use EmptyDir Volumes" }
func (l *CKADVolumeEmptyDirLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADVolumeEmptyDirLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADVolumeEmptyDirLab) Cert() Cert             { return CertCKAD }
func (l *CKADVolumeEmptyDirLab) DomainWeight() int      { return 20 }
func (l *CKADVolumeEmptyDirLab) EstimatedTime() int     { return 15 }
func (l *CKADVolumeEmptyDirLab) Tags() []string {
	return []string{"emptydir", "volumes", "storage"}
}

func (l *CKADVolumeEmptyDirLab) Description() string {
	return `A pod needs to share data between two containers using an EmptyDir volume.
The main container writes data, and the sidecar reads it.

Your task: Configure the pod with an EmptyDir volume shared between both containers.`
}

func (l *CKADVolumeEmptyDirLab) Hints() []string {
	return []string{
		"EmptyDir volumes are created when a pod is assigned to a node",
		"Both containers need to mount the same volume",
		"Use different mountPaths for each container if needed",
	}
}

func (l *CKADVolumeEmptyDirLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADVolumeEmptyDirLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: shared-data
  labels:
    app: shared-data
spec:
  containers:
  - name: writer
    image: busybox:1.36
    command: ['sh', '-c', 'echo "hello" > /data/message.txt && sleep 3600']
  - name: reader
    image: busybox:1.36
    command: ['sh', '-c', 'cat /data/message.txt && sleep 3600']`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADVolumeEmptyDirLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "shared-data",
		"-o", "jsonpath={.spec.volumes[*].emptyDir}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no EmptyDir volume found")
	}
	return nil
}

func (l *CKADVolumeEmptyDirLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod to add EmptyDir", Command: "kubectl edit pod shared-data"},
		{Description: "Add volumes section", Command: "Add volumes with emptyDir and mount in both containers"},
		{Description: "Verify volume is shared", Command: "kubectl get pod shared-data -o yaml"},
	}
}
