package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&NamespaceFinalizerStuck{})
}

type NamespaceFinalizerStuck struct {
	BaseLab
}

func (l *NamespaceFinalizerStuck) ID() string             { return "namespace_finalizer_stuck2" }
func (l *NamespaceFinalizerStuck) Title() string          { return "Namespace Stuck in Terminating" }
func (l *NamespaceFinalizerStuck) Category() Category     { return CategoryControlPlane }
func (l *NamespaceFinalizerStuck) Difficulty() Difficulty { return DifficultyHard }
func (l *NamespaceFinalizerStuck) EstimatedTime() int     { return 20 }
func (l *NamespaceFinalizerStuck) Tags() []string {
	return []string{"namespaces", "finalizers", "deletion"}
}

func (l *NamespaceFinalizerStuck) Description() string {
	return `A namespace is stuck in Terminating state because of a stuck finalizer.
Remove the finalizer to allow the namespace to be deleted.`
}

func (l *NamespaceFinalizerStuck) Hints() []string {
	return []string{
		"Check namespace finalizers",
		"Look at the namespace status",
		"Remove the finalizer using kubectl patch",
	}
}

func (l *NamespaceFinalizerStuck) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NamespaceFinalizerStuck) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: stuck-ns
  finalizers:
  - foregroundDeletion`
	if err := kubectlApply(ctx, kubeconfigPath, manifest); err != nil {
		return err
	}
	_, _ = kubectl(ctx, kubeconfigPath, "delete", "namespace", "stuck-ns")
	return nil
}

func (l *NamespaceFinalizerStuck) Verify(ctx context.Context, kubeconfigPath string) error {
	_, err := kubectl(ctx, kubeconfigPath, "get", "namespace", "stuck-ns")
	if err != nil {
		return nil // namespace doesn't exist = success
	}
	return fmt.Errorf("namespace still exists")
}

func (l *NamespaceFinalizerStuck) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check namespace", Command: "kubectl get ns stuck-ns"},
		{Description: "Check finalizers", Command: "kubectl get ns stuck-ns -o jsonpath='{.spec.finalizers}'"},
		{Description: "Remove finalizer", Command: "kubectl patch ns stuck-ns -p '{\"metadata\":{\"finalizers\":null}}' --type=merge"},
	}
}
