package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADMultiContainerSidecarLab{})
}

type CKADMultiContainerSidecarLab struct {
	BaseLab
}

func (l *CKADMultiContainerSidecarLab) ID() string             { return "ckad_multi_container_sidecar" }
func (l *CKADMultiContainerSidecarLab) Title() string          { return "Configure Sidecar Pattern" }
func (l *CKADMultiContainerSidecarLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADMultiContainerSidecarLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADMultiContainerSidecarLab) Cert() Cert             { return CertCKAD }
func (l *CKADMultiContainerSidecarLab) DomainWeight() int      { return 20 }
func (l *CKADMultiContainerSidecarLab) EstimatedTime() int     { return 20 }
func (l *CKADMultiContainerSidecarLab) Tags() []string {
	return []string{"sidecar", "multi-container", "patterns"}
}

func (l *CKADMultiContainerSidecarLab) Description() string {
	return `A web application pod is running but needs a sidecar container to
collect and forward logs to a central logging system.

Your task: Add a sidecar container to the existing pod that watches
the shared log directory and outputs log contents.`
}

func (l *CKADMultiContainerSidecarLab) Hints() []string {
	return []string{
		"Use a shared EmptyDir volume between containers",
		"The sidecar should mount the same log directory",
		"Use a lightweight image like busybox for the sidecar",
	}
}

func (l *CKADMultiContainerSidecarLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADMultiContainerSidecarLab) Break(ctx context.Context, kubeconfigPath string) error {
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
    - name: log-volume
      mountPath: /var/log/nginx
  volumes:
  - name: log-volume
    emptyDir: {}`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADMultiContainerSidecarLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "webapp",
		"-o", "jsonpath={.spec.containers[*].name}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	names := strings.Split(strings.TrimSpace(output), " ")
	if len(names) < 2 {
		return fmt.Errorf("pod has only %d container(s), expected at least 2", len(names))
	}
	return nil
}

func (l *CKADMultiContainerSidecarLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod to add sidecar", Command: "kubectl edit pod webapp"},
		{Description: "Add sidecar container", Command: "Add a container named 'log-reader' using busybox that tails /var/log/nginx"},
		{Description: "Verify sidecar is running", Command: "kubectl get pod webapp"},
	}
}
