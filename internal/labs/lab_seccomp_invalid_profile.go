package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&SeccompInvalidProfileLab{}) }

type SeccompInvalidProfileLab struct{ BaseLab }

func (l *SeccompInvalidProfileLab) ID() string          { return "seccomp_invalid_profile" }
func (l *SeccompInvalidProfileLab) Title() string        { return "Pod Rejected — Invalid seccomp Profile" }
func (l *SeccompInvalidProfileLab) Category() Category   { return CategorySecurity }
func (l *SeccompInvalidProfileLab) Difficulty() Difficulty { return DifficultyHard }
func (l *SeccompInvalidProfileLab) EstimatedTime() int   { return 20 }
func (l *SeccompInvalidProfileLab) Tags() []string {
	return []string{"security", "seccomp", "profile", "admission"}
}
func (l *SeccompInvalidProfileLab) Description() string {
return `A pod 'seccomp-worker' in namespace 'strict-ns' (same namespace as
the runAsNonRoot lab) fails with: violates PodSecurity
"restricted:latest": seccompProfile (seccomp profile must be
RuntimeDefault or Localhost).

The pod sets seccompProfile.type: Unconfined, which violates the
restricted profile.

Your task: Fix the seccomp profile to use RuntimeDefault instead of
Unconfined, then recreate the pod.`
}
func (l *SeccompInvalidProfileLab) Hints() []string {
	return []string{
		"Check: kubectl get pod seccomp-worker -n strict-ns -o yaml | grep seccomp",
		"Unconfined is not allowed under the restricted profile",
		"Change type to RuntimeDefault (or Localhost with a valid path)",
		"Delete and recreate the pod",
	}
}

func (l *SeccompInvalidProfileLab) Break(ctx context.Context, kp string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: seccomp-worker
  namespace: strict-ns
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    seccompProfile:
      type: Unconfined
  containers:
  - name: worker
    image: busybox:1.36
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop: ["ALL"]
    command: ["sh","-c","id; while true; do sleep 5; done"]
`
	return kubectlApply(ctx, kp, pod)
}

func (l *SeccompInvalidProfileLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *SeccompInvalidProfileLab) Verify(ctx context.Context, kp string) error {
	phase, _ := kubectl(ctx, kp, "get", "pod", "seccomp-worker", "-n", "strict-ns", "-o", "jsonpath={.status.phase}")
	if phase != "Running" {
		return fmt.Errorf("pod not running (phase: %s)", phase)
	}
	seccomp, _ := kubectl(ctx, kp, "get", "pod", "seccomp-worker", "-n", "strict-ns", "-o",
		"jsonpath={.spec.securityContext.seccompProfile.type}")
	if seccomp == "Unconfined" {
		return fmt.Errorf("seccomp profile is still Unconfined")
	}
	return nil
}

func (l *SeccompInvalidProfileLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check the error", Command: "kubectl describe pod seccomp-worker -n strict-ns | tail -10", Notes: "Forbidden: seccompProfile type must be RuntimeDefault or Localhost"},
		{Description: "Fix seccomp profile", Command: "kubectl delete pod seccomp-worker -n strict-ns", Notes: "Pod spec is immutable"},
		{Description: "Recreate with RuntimeDefault", Command: `kubectl run seccomp-worker -n strict-ns --image=busybox:1.36 --restart=Never --overrides='{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"seccompProfile":{"type":"RuntimeDefault"}},"containers":[{"name":"seccomp-worker","image":"busybox:1.36","securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}},"command":["sh","-c","id; while true; do sleep 5; done"]}]}}'`, Notes: "RuntimeDefault satisfies the restricted profile requirement"},
		{Description: "Verify", Command: "kubectl get pod seccomp-worker -n strict-ns -o jsonpath='{.spec.securityContext.seccompProfile.type}'", Notes: "Shows RuntimeDefault"},
	}
}
