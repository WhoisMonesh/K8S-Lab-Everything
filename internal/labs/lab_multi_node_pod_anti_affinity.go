package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&MultiNodePodAntiAffinityLab{})
}

type MultiNodePodAntiAffinityLab struct {
	BaseLab
}

func (l *MultiNodePodAntiAffinityLab) ID() string {
	return "multi_node_pod_anti_affinity"
}

func (l *MultiNodePodAntiAffinityLab) Title() string {
	return "Pod Anti-Affinity Scheduling"
}

func (l *MultiNodePodAntiAffinityLab) Category() Category {
	return CategoryWorkloadsScheduling
}

func (l *MultiNodePodAntiAffinityLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *MultiNodePodAntiAffinityLab) Description() string {
	return `A deployment should ensure that replica pods run on different nodes
for fault tolerance, but the anti-affinity rule is too strict and prevents
scheduling entirely.

Your task: Fix the pod anti-affinity configuration so replicas are distributed
across different nodes without preventing scheduling.`
}

func (l *MultiNodePodAntiAffinityLab) Hints() []string {
	return []string{
		"Check the current anti-affinity rule",
		"requiredDuringSchedulingIgnoredDuringExecution is strict",
		"Use preferredDuringSchedulingIgnoredDuringExecution for soft preference",
	}
}

func (l *MultiNodePodAntiAffinityLab) EstimatedTime() int {
	return 20
}

func (l *MultiNodePodAntiAffinityLab) Tags() []string {
	return []string{"pod-anti-affinity", "scheduling", "multi-node", "fault-tolerance"}
}

func (l *MultiNodePodAntiAffinityLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

// ClusterSpec declares a multi-worker cluster so node scheduling/scaling
// scenarios are real on kind.
func (l *MultiNodePodAntiAffinityLab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           2,
	}
}

func (l *MultiNodePodAntiAffinityLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: anti-affinity-app
  namespace: default
spec:
  replicas: 4
  selector:
    matchLabels:
      app: anti-affinity-app
  template:
    metadata:
      labels:
        app: anti-affinity-app
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - anti-affinity-app
            topologyKey: kubernetes.io/hostname
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo anti-affinity; sleep 15; done']
        resources:
          limits:
            memory: 32Mi
            cpu: 50m
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *MultiNodePodAntiAffinityLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=anti-affinity-app",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("checking pods: %w", err)
	}

	phases := strings.Fields(output)
	running := 0
	for _, p := range phases {
		if p == "Running" {
			running++
		}
	}

	if running < 4 {
		return nil
	}

	return fmt.Errorf("all pods are running (expected scheduling failure)")
}

func (l *MultiNodePodAntiAffinityLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "anti-affinity-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("checking deployment: %w", err)
	}

	ready := strings.TrimSpace(output)
	if ready != "4" {
		return fmt.Errorf("deployment not ready (ready: %s, expected: 4)", ready)
	}

	nodes, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=anti-affinity-app",
		"-o", "jsonpath={range .items[*]}{.spec.nodeName}{'\n'}{end}")
	if err != nil {
		return fmt.Errorf("checking pod nodes: %w", err)
	}

	nodeCount := make(map[string]int)
	for _, n := range strings.Split(strings.TrimSpace(nodes), "\n") {
		if n != "" {
			nodeCount[n]++
		}
	}

	if len(nodeCount) < 2 {
		return fmt.Errorf("pods not distributed across nodes")
	}

	return nil
}

func (l *MultiNodePodAntiAffinityLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pods -l app=anti-affinity-app",
			Notes:       "Pods may be stuck in Pending",
		},
		{
			Description: "Fix: Change to preferred anti-affinity",
			Command:     `kubectl patch deploy anti-affinity-app --type='json' -p='[{"op":"replace","path":"/spec/template/spec/affinity/podAntiAffinity","value":{"preferredDuringSchedulingIgnoredDuringExecution":[{"weight":100,"podAffinityTerm":{"labelSelector":{"matchExpressions":[{"key":"app","operator":"In","values":["anti-affinity-app"]}]},"topologyKey":"kubernetes.io/hostname"}}]}}]'`,
			Notes:       "Use preferred scheduling instead of required",
		},
		{
			Description: "Verify all pods are running",
			Command:     "kubectl rollout status deploy/anti-affinity-app",
			Notes:       "All 4 replicas should be ready",
		},
		{
			Description: "Check pod distribution",
			Command:     "kubectl get pods -l app=anti-affinity-app -o wide",
			Notes:       "Pods should be on different nodes when possible",
		},
	}
}
