package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&LimitRangeExceededLab{}) }

type LimitRangeExceededLab struct{ BaseLab }

func (l *LimitRangeExceededLab) ID() string          { return "limitrange_exceeded" }
func (l *LimitRangeExceededLab) Title() string        { return "Pod Rejected by LimitRange" }
func (l *LimitRangeExceededLab) Category() Category   { return CategoryScheduling }
func (l *LimitRangeExceededLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *LimitRangeExceededLab) EstimatedTime() int   { return 15 }
func (l *LimitRangeExceededLab) Tags() []string {
	return []string{"limitrange", "resource", "quota", "scheduling"}
}
func (l *LimitRangeExceededLab) Description() string {
	return `A pod named 'big-app' in namespace 'limitns' is stuck Pending with
resourcequota/message: exceeded quota — but actually the LimitRange
max CPU is 500m and the pod requests 1000m.

Your task: Fix the LimitRange max to allow the pod (increase max to 2),
then delete and recreate the pod.`
}
func (l *LimitRangeExceededLab) Hints() []string {
	return []string{
		"kubectl describe pod big-app shows LimitRange admission error",
		"Check LimitRange max: kubectl get limitrange -n limitns -o yaml",
		"Increase max CPU to 2000m so the pod's 1000m request fits",
		"Delete the old pending pod after fixing the LimitRange",
	}
}

func (l *LimitRangeExceededLab) Break(ctx context.Context, kp string) error {
	if _, err := kubectl(ctx, kp, "create", "ns", "limitns"); err != nil {
		return err
	}
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: limitns
`
	lr := `apiVersion: v1
kind: LimitRange
metadata:
  name: cpu-limit
  namespace: limitns
spec:
  limits:
  - type: Container
    max:
      cpu: "500m"
    default:
      cpu: "250m"
`
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: big-app
  namespace: limitns
spec:
  containers:
  - name: app
    image: nginx:1.27-alpine
    resources:
      requests:
        cpu: "1000m"
      limits:
        cpu: "1000m"
`
	kubectlApply(ctx, kp, ns)
	kubectlApply(ctx, kp, lr)
	return kubectlApply(ctx, kp, pod)
}

func (l *LimitRangeExceededLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *LimitRangeExceededLab) Verify(ctx context.Context, kp string) error {
	maxCPU, _ := kubectl(ctx, kp, "get", "limitrange", "cpu-limit", "-n", "limitns", "-o",
		"jsonpath={.spec.limits[0].max.cpu}")
	if maxCPU == "500m" || maxCPU == "" {
		return fmt.Errorf("LimitRange max CPU is still 500m")
	}
	phase, _ := kubectl(ctx, kp, "get", "pod", "big-app", "-n", "limitns", "-o",
		"jsonpath={.status.phase}")
	if phase != "Running" {
		return fmt.Errorf("pod big-app not running (phase: %s)", phase)
	}
	return nil
}

func (l *LimitRangeExceededLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Diagnose the pending pod", Command: "kubectl describe pod big-app -n limitns | tail -10", Notes: "Shows LimitRange admission error"},
		{Description: "Check LimitRange", Command: "kubectl get limitrange cpu-limit -n limitns -o yaml", Notes: "max.cpu = 500m but pod requests 1000m"},
		{Description: "Fix the LimitRange max", Command: `kubectl patch limitrange cpu-limit -n limitns -p '{"spec":{"limits":[{"type":"Container","max":{"cpu":"2"},"default":{"cpu":"250m"}}]}}'`, Notes: "Increase max to 2 CPU"},
		{Description: "Delete and recreate the pod", Command: "kubectl delete pod big-app -n limitns && kubectl apply -f <fixed-pod.yaml>", Notes: "Old pod's admission decision is cached; recreating picks up new LimitRange"},
	}
}
