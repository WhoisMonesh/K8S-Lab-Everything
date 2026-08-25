package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&NamespaceFinalizerStuckLab{}) }

type NamespaceFinalizerStuckLab struct{ BaseLab }

func (l *NamespaceFinalizerStuckLab) ID() string          { return "namespace_finalizer_stuck" }
func (l *NamespaceFinalizerStuckLab) Title() string        { return "Namespace Stuck in Terminating" }
func (l *NamespaceFinalizerStuckLab) Category() Category   { return CategoryControlPlane }
func (l *NamespaceFinalizerStuckLab) Difficulty() Difficulty { return DifficultyHard }
func (l *NamespaceFinalizerStuckLab) EstimatedTime() int   { return 20 }
func (l *NamespaceFinalizerStuckLab) Tags() []string {
	return []string{"namespace", "finalizer", "control-plane"}
}
func (l *NamespaceFinalizerStuckLab) Description() string {
	return `A namespace 'doomed' has been deleted but is stuck in Terminating
status indefinitely. This happens when finalizers prevent the
namespace from being garbage collected.

Your task: Remove the finalizers from the namespace so it can
complete termination. You'll need to use a JSON patch or edit the
namespace to strip the finalizers list.`
}
func (l *NamespaceFinalizerStuckLab) Hints() []string {
	return []string{
		"Check: kubectl get ns doomed -o yaml | grep finalizers",
		"Finalizers block deletion — strip them to unstick",
		"Use: kubectl patch ns doomed -p '{\"metadata\":{\"finalizers\":null}}' --type=merge",
		"Or: kubectl get ns doomed -o json | jq '.metadata.finalizers=null' | kubectl replace --raw /api/v1/namespaces/doomed -f -",
	}
}

func (l *NamespaceFinalizerStuckLab) Break(ctx context.Context, kp string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: doomed
  finalizers:
  - foreground-deletion
  - kubernetes
`
	if err := kubectlApply(ctx, kp, ns); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	// Start the deletion — the finalizers will block it
	_, _ = kubectl(ctx, kp, "delete", "ns", "doomed")
	return nil
}

func (l *NamespaceFinalizerStuckLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *NamespaceFinalizerStuckLab) Verify(ctx context.Context, kp string) error {
	out, err := kubectl(ctx, kp, "get", "ns", "doomed", "-o", "jsonpath={.status.phase}")
	if err != nil || strings.TrimSpace(out) == "" {
		return nil // namespace is gone — success
	}
	if strings.Contains(out, "Terminating") {
		return fmt.Errorf("namespace still stuck in Terminating")
	}
	return nil
}

func (l *NamespaceFinalizerStuckLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Confirm namespace is stuck", Command: "kubectl get ns doomed", Notes: "Shows Terminating status"},
		{Description: "Check finalizers", Command: "kubectl get ns doomed -o jsonpath='{.metadata.finalizers}'", Notes: "Shows [\"foreground-deletion\",\"kubernetes\"]"},
		{Description: "Remove finalizers", Command: `kubectl patch ns doomed -p '{"metadata":{"finalizers":null}}' --type=merge`, Notes: "Strips all finalizers, allowing garbage collection"},
		{Description: "Verify deletion completes", Command: "kubectl get ns doomed", Notes: "Returns 'Error from server (NotFound)' — namespace is gone"},
	}
}
