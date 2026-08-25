package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&RunAsNonRootRejectedLab{}) }

type RunAsNonRootRejectedLab struct{ BaseLab }

func (l *RunAsNonRootRejectedLab) ID() string             { return "runasnonroot_rejected" }
func (l *RunAsNonRootRejectedLab) Title() string          { return "Pod Rejected — runAsNonRoot Violation" }
func (l *RunAsNonRootRejectedLab) Category() Category     { return CategorySecurity }
func (l *RunAsNonRootRejectedLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *RunAsNonRootRejectedLab) EstimatedTime() int     { return 15 }
func (l *RunAsNonRootRejectedLab) Tags() []string {
	return []string{"security", "securitycontext", "runasnonroot", "admission"}
}
func (l *RunAsNonRootRejectedLab) Description() string {
	return `A pod 'secure-worker' in namespace 'strict-ns' fails to create with
error: pods "secure-worker" is forbidden: violates PodSecurity
"restricted:latest": runAsNonRoot (pod must not run as root).

The PodSecurity admission enforces restricted profile on the namespace,
but the pod doesn't set runAsNonRoot: true or runAsUser to a non-zero
value.

Your task: Fix the pod's securityContext to satisfy the restricted
profile by setting runAsNonRoot: true and runAsUser: 1000.`
}
func (l *RunAsNonRootRejectedLab) Hints() []string {
	return []string{
		"Check: kubectl describe ns strict-ns — look for pod-security.kubernetes.io/enforce",
		"restricted profile requires runAsNonRoot: true",
		"Also set runAsUser to a non-zero UID (e.g. 1000)",
		"Delete and recreate the pod with the corrected securityContext",
	}
}

func (l *RunAsNonRootRejectedLab) Break(ctx context.Context, kp string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: strict-ns
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/enforce-version: latest
    pod-security.kubernetes.io/warn: restricted
`
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: secure-worker
  namespace: strict-ns
spec:
  securityContext:
    runAsUser: 0
  containers:
  - name: worker
    image: busybox:1.36
    command: ["sh","-c","id; while true; do sleep 5; done"]
`
	kubectlApply(ctx, kp, ns)
	return kubectlApply(ctx, kp, pod)
}

func (l *RunAsNonRootRejectedLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *RunAsNonRootRejectedLab) Verify(ctx context.Context, kp string) error {
	// Check pod exists and is Running
	phase, err := kubectl(ctx, kp, "get", "pod", "secure-worker", "-n", "strict-ns", "-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("pod not found or error: %w", err)
	}
	if phase != "Running" {
		return fmt.Errorf("pod not running (phase: %s)", phase)
	}
	// Verify runAsNonRoot is set
	runAsRoot, _ := kubectl(ctx, kp, "get", "pod", "secure-worker", "-n", "strict-ns", "-o",
		"jsonpath={.spec.securityContext.runAsNonRoot}")
	if runAsRoot != "true" {
		return fmt.Errorf("runAsNonRoot is not true")
	}
	return nil
}

func (l *RunAsNonRootRejectedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check namespace enforcement", Command: "kubectl get ns strict-ns --show-labels | grep pod-security", Notes: "Enforces restricted profile"},
		{Description: "Check pod failure", Command: "kubectl describe pod secure-worker -n strict-ns | tail -10", Notes: "Forbidden: runAsNonRoot violation"},
		{Description: "Delete and recreate with correct securityContext", Command: "kubectl delete pod secure-worker -n strict-ns", Notes: "Pod is immutable for securityContext"},
		{Description: "Recreate with non-root user", Command: `kubectl run secure-worker -n strict-ns --image=busybox:1.36 --restart=Never --overrides='{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000},"containers":[{"name":"secure-worker","image":"busybox:1.36","command":["sh","-c","id; while true; do sleep 5; done"]}]}}'`, Notes: "runAsUser:1000 + runAsNonRoot:true satisfies restricted profile"},
		{Description: "Verify", Command: "kubectl get pod secure-worker -n strict-ns && kubectl logs secure-worker -n strict-ns --tail=1", Notes: "Pod Running, id shows uid=1000"},
	}
}
