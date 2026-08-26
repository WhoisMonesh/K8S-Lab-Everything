package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&RunAsNonRootRejected{})
}

type RunAsNonRootRejected struct {
	BaseLab
}

func (l *RunAsNonRootRejected) ID() string             { return "runasnonroot_rejected2" }
func (l *RunAsNonRootRejected) Title() string          { return "Pod Rejected - runAsNonRoot Violation" }
func (l *RunAsNonRootRejected) Category() Category     { return CategorySecurity }
func (l *RunAsNonRootRejected) Difficulty() Difficulty { return DifficultyMedium }
func (l *RunAsNonRootRejected) EstimatedTime() int     { return 15 }
func (l *RunAsNonRootRejected) Tags() []string {
	return []string{"security", "runasnonroot", "admission"}
}

func (l *RunAsNonRootRejected) Description() string {
	return `A pod is being rejected because it tries to run as root when runAsNonRoot is enforced.
Fix the pod security context to run as a non-root user.`
}

func (l *RunAsNonRootRejected) Hints() []string {
	return []string{
		"Check the pod securityContext",
		"Look for runAsNonRoot setting",
		"Set runAsUser to a non-zero value",
	}
}

func (l *RunAsNonRootRejected) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *RunAsNonRootRejected) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Pod
metadata:
  name: root-pod
spec:
  securityContext:
    runAsNonRoot: true
  containers:
  - name: nginx
    image: nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *RunAsNonRootRejected) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "root-pod",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("pod not running: %s", output)
	}
	return nil
}

func (l *RunAsNonRootRejected) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod securityContext", Command: "kubectl get pod root-pod -o jsonpath='{.spec.securityContext}'"},
		{Description: "Fix securityContext", Command: "kubectl edit pod root-pod"},
		{Description: "Add runAsUser", Command: "Add runAsUser: 1000 to securityContext"},
	}
}
