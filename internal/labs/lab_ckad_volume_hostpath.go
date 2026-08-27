package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADVolumeHostPathLab{})
}

type CKADVolumeHostPathLab struct {
	BaseLab
}

func (l *CKADVolumeHostPathLab) ID() string             { return "ckad_volume_hostpath" }
func (l *CKADVolumeHostPathLab) Title() string          { return "Use HostPath Volumes" }
func (l *CKADVolumeHostPathLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADVolumeHostPathLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADVolumeHostPathLab) Cert() Cert             { return CertCKAD }
func (l *CKADVolumeHostPathLab) DomainWeight() int      { return 20 }
func (l *CKADVolumeHostPathLab) EstimatedTime() int     { return 15 }
func (l *CKADVolumeHostPathLab) Tags() []string {
	return []string{"hostpath", "volumes", "storage"}
}

func (l *CKADVolumeHostPathLab) Description() string {
	return `A monitoring agent needs to read system logs from the host node.
Configure the pod to mount the host's /var/log directory.

Your task: Add a HostPath volume to the pod to access host logs.`
}

func (l *CKADVolumeHostPathLab) Hints() []string {
	return []string{
		"HostPath volumes mount a file or directory from the host node",
		"Use type: Directory for mounting existing directories",
		"Be careful with security implications of HostPath",
	}
}

func (l *CKADVolumeHostPathLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADVolumeHostPathLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: log-reader
  labels:
    app: log-reader
spec:
  containers:
  - name: reader
    image: busybox:1.36
    command: ['sh', '-c', 'while true; do sleep 3600; done']
    volumeMounts: []`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADVolumeHostPathLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "log-reader",
		"-o", "jsonpath={.spec.volumes[*].hostPath}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no HostPath volume found")
	}
	if !strings.Contains(output, "/var/log") {
		return fmt.Errorf("HostPath not mounting /var/log")
	}
	return nil
}

func (l *CKADVolumeHostPathLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod to add HostPath", Command: "kubectl edit pod log-reader"},
		{Description: "Add HostPath volume", Command: "Add hostPath volume with path /var/log"},
		{Description: "Verify mount", Command: "kubectl exec log-reader -- ls /var/log"},
	}
}
