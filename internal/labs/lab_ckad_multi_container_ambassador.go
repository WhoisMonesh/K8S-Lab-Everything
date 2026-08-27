package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADMultiContainerAmbassadorLab{})
}

type CKADMultiContainerAmbassadorLab struct {
	BaseLab
}

func (l *CKADMultiContainerAmbassadorLab) ID() string             { return "ckad_multi_container_ambassador" }
func (l *CKADMultiContainerAmbassadorLab) Title() string          { return "Configure Ambassador Pattern" }
func (l *CKADMultiContainerAmbassadorLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADMultiContainerAmbassadorLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKADMultiContainerAmbassadorLab) Cert() Cert             { return CertCKAD }
func (l *CKADMultiContainerAmbassadorLab) DomainWeight() int      { return 20 }
func (l *CKADMultiContainerAmbassadorLab) EstimatedTime() int     { return 25 }
func (l *CKADMultiContainerAmbassadorLab) Tags() []string {
	return []string{"ambassador", "multi-container", "proxy", "patterns"}
}

func (l *CKADMultiContainerAmbassadorLab) Description() string {
	return `A microservice needs to communicate with an external database through
an ambassador proxy that handles connection pooling and TLS termination.

Your task: Add an ambassador container that proxies connections from the
main application container to the external database service.`
}

func (l *CKADMultiContainerAmbassadorLab) Hints() []string {
	return []string{
		"The ambassador container should act as a proxy",
		"Use shared localhost networking between containers",
		"Consider using socat or nginx as the ambassador",
	}
}

func (l *CKADMultiContainerAmbassadorLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADMultiContainerAmbassadorLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: myapp
  labels:
    app: myapp
spec:
  containers:
  - name: myapp
    image: busybox:1.36
    command: ['sh', '-c', 'while true; do sleep 3600; done']`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADMultiContainerAmbassadorLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "myapp",
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

func (l *CKADMultiContainerAmbassadorLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod to add ambassador", Command: "kubectl edit pod myapp"},
		{Description: "Add ambassador container", Command: "Add a container using socat or nginx that proxies connections"},
		{Description: "Verify ambassador is running", Command: "kubectl get pod myapp"},
	}
}
