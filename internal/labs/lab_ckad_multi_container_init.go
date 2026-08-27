package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADMultiContainerInitLab{})
}

type CKADMultiContainerInitLab struct {
	BaseLab
}

func (l *CKADMultiContainerInitLab) ID() string             { return "ckad_multi_container_init" }
func (l *CKADMultiContainerInitLab) Title() string          { return "Configure Init Container Pattern" }
func (l *CKADMultiContainerInitLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADMultiContainerInitLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADMultiContainerInitLab) Cert() Cert             { return CertCKAD }
func (l *CKADMultiContainerInitLab) DomainWeight() int      { return 20 }
func (l *CKADMultiContainerInitLab) EstimatedTime() int     { return 20 }
func (l *CKADMultiContainerInitLab) Tags() []string {
	return []string{"init-container", "multi-container", "patterns"}
}

func (l *CKADMultiContainerInitLab) Description() string {
	return `A web application needs to wait for a configuration file to be
ready before starting. The init container should fetch the config
from a service before the main container starts.

Your task: Add an init container that downloads configuration before
the main application starts.`
}

func (l *CKADMultiContainerInitLab) Hints() []string {
	return []string{
		"Init containers run before the main container starts",
		"Use a shared volume to pass data between init and main containers",
		"The init container should complete successfully before the main container starts",
	}
}

func (l *CKADMultiContainerInitLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADMultiContainerInitLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: webapp
  labels:
    app: webapp
spec:
  containers:
  - name: webapp
    image: nginx:alpine
    volumeMounts:
    - name: config-volume
      mountPath: /app/config
  volumes:
  - name: config-volume
    emptyDir: {}`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADMultiContainerInitLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "webapp",
		"-o", "jsonpath={.spec.initContainers[*].name}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no init containers found")
	}
	return nil
}

func (l *CKADMultiContainerInitLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod to add init container", Command: "kubectl edit pod webapp"},
		{Description: "Add init container", Command: "Add initContainers with a busybox container that waits or fetches config"},
		{Description: "Verify init container ran", Command: "kubectl get pod webapp -o yaml | grep initContainers"},
	}
}
