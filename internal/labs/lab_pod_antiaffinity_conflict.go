package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&PodAntiAffinityConflictLab{}) }

type PodAntiAffinityConflictLab struct{ BaseLab }

func (l *PodAntiAffinityConflictLab) ID() string { return "pod_antiaffinity_conflict" }
func (l *PodAntiAffinityConflictLab) Title() string {
	return "Deployment Can't Schedule Due to Anti-Affinity"
}
func (l *PodAntiAffinityConflictLab) Category() Category     { return CategoryScheduling }
func (l *PodAntiAffinityConflictLab) Difficulty() Difficulty { return DifficultyHard }
func (l *PodAntiAffinityConflictLab) EstimatedTime() int     { return 20 }
func (l *PodAntiAffinityConflictLab) Tags() []string {
	return []string{"antiaffinity", "scheduling", "replicas"}
}
func (l *PodAntiAffinityConflictLab) Description() string {
	return `A deployment 'distributed-cache' needs 3 replicas but only 1 ever
runs. The others stay Pending due to a podAntiAffinity rule that
requires each replica to be on a DIFFERENT node — but the cluster
only has 2 nodes (one is a control-plane with a NoSchedule taint).

Your task: Remove the podAntiAffinity requirement so all 3 replicas
can schedule on the same node. Alternatively, you could fix the
control-plane taint, but removing the anti-affinity is simpler.`
}
func (l *PodAntiAffinityConflictLab) Hints() []string {
	return []string{
		"Check: kubectl get nodes -o wide — how many worker nodes?",
		"Check: kubectl describe deploy distributed-cache — look for anti-affinity",
		"requiredDuringSchedulingIgnoredDuringExecution is strict — switch to preferredDuringSchedulingIgnoredDuringExecution",
		"Or remove the podAntiAffinity block entirely",
	}
}

func (l *PodAntiAffinityConflictLab) Break(ctx context.Context, kp string) error {
	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: distributed-cache
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cache
  template:
    metadata:
      labels:
        app: cache
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - cache
            topologyKey: kubernetes.io/hostname
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
`
	return kubectlApply(ctx, kp, deploy)
}

func (l *PodAntiAffinityConflictLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodAntiAffinityConflictLab) Verify(ctx context.Context, kp string) error {
	ready, _ := kubectl(ctx, kp, "get", "deploy", "distributed-cache", "-o",
		"jsonpath={.status.readyReplicas}")
	if ready != "3" {
		return fmt.Errorf("not all replicas ready (ready: %s)", ready)
	}
	// Verify the anti-affinity was relaxed, not removed entirely
	aff, _ := kubectl(ctx, kp, "get", "deploy", "distributed-cache", "-o",
		"jsonpath={.spec.template.spec.affinity.podAntiAffinity}")
	if aff == "" {
		return nil // completely removed — still valid
	}
	if strings.Contains(aff, "requiredDuring") {
		return fmt.Errorf("still has requiredDuringScheduling anti-affinity")
	}
	return nil
}

func (l *PodAntiAffinityConflictLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check node count", Command: "kubectl get nodes --show-labels | grep -v master", Notes: "Only 1 worker node available"},
		{Description: "Check scheduling", Command: "kubectl describe deploy distributed-cache | grep -A10 Events", Notes: "Pods Pending due to unsatisfiable anti-affinity"},
		{Description: "Patch to remove anti-affinity", Command: `kubectl patch deploy distributed-cache --type=json -p='[{"op":"remove","path":"/spec/template/spec/affinity"}]'`, Notes: "Removes the strict pod-to-different-node requirement"},
		{Description: "Verify", Command: "kubectl rollout status deploy/distributed-cache && kubectl get pods -l app=cache -o wide", Notes: "All 3 replicas Running, possibly on the same node"},
	}
}
