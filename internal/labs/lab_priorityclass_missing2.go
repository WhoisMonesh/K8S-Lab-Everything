package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PriorityClassMissing{})
}

type PriorityClassMissing struct {
	BaseLab
}

func (l *PriorityClassMissing) ID() string             { return "priorityclass_missing2" }
func (l *PriorityClassMissing) Title() string          { return "Pod Uses Nonexistent PriorityClass" }
func (l *PriorityClassMissing) Category() Category     { return CategoryScheduling }
func (l *PriorityClassMissing) Difficulty() Difficulty { return DifficultyMedium }
func (l *PriorityClassMissing) EstimatedTime() int     { return 15 }
func (l *PriorityClassMissing) Tags() []string         { return []string{"scheduling", "priority", "class"} }

func (l *PriorityClassMissing) Description() string {
	return `A pod references a PriorityClass that doesn't exist.
Create the missing PriorityClass or fix the pod reference.`
}

func (l *PriorityClassMissing) Hints() []string {
	return []string{
		"Check available PriorityClasses",
		"Look at the pod priorityClassName",
		"Create the missing PriorityClass",
	}
}

func (l *PriorityClassMissing) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PriorityClassMissing) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Pod
metadata:
  name: priority-pod
spec:
  priorityClassName: high-priority
  containers:
  - name: nginx
    image: nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PriorityClassMissing) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "priority-pod",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("pod not running: %s", output)
	}
	return nil
}

func (l *PriorityClassMissing) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check PriorityClasses", Command: "kubectl get priorityclasses"},
		{Description: "Create PriorityClass", Command: "kubectl create priorityclass high-priority --value=1000000 --global-default=false"},
	}
}
