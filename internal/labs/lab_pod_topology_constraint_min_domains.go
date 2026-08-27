package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodTopologyConstraintMinDomainsLab{})
}

type PodTopologyConstraintMinDomainsLab struct {
	BaseLab
}

func (l *PodTopologyConstraintMinDomainsLab) ID() string {
	return "pod_topology_constraint_min_domains"
}

func (l *PodTopologyConstraintMinDomainsLab) Title() string {
	return "Topology Constraint minDomains Too High"
}

func (l *PodTopologyConstraintMinDomainsLab) Category() Category {
	return CategoryScheduling
}

func (l *PodTopologyConstraintMinDomainsLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *PodTopologyConstraintMinDomainsLab) Description() string {
	return `A Deployment 'distributed-app' has a topology spread constraint with
minDomains=5 but the cluster only has 3 nodes. Pods remain Pending
because the constraint cannot be satisfied.

Your task: Fix the minDomains value to match cluster capacity.`
}

func (l *PodTopologyConstraintMinDomainsLab) Hints() []string {
	return []string{
		"Check the pod topology spread constraints",
		"minDomains must be <= number of matching nodes",
		"Reduce minDomains to match available nodes",
	}
}

func (l *PodTopologyConstraintMinDomainsLab) EstimatedTime() int {
	return 15
}

func (l *PodTopologyConstraintMinDomainsLab) Tags() []string {
	return []string{"topology", "spread", "constraint", "scheduling"}
}

func (l *PodTopologyConstraintMinDomainsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodTopologyConstraintMinDomainsLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: distributed-app
  namespace: default
spec:
  replicas: 6
  selector:
    matchLabels:
      app: distributed-app
  template:
    metadata:
      labels:
        app: distributed-app
    spec:
      topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: kubernetes.io/hostname
        whenUnsatisfiable: DoNotSchedule
        labelSelector:
          matchLabels:
            app: distributed-app
        minDomains: 5
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'sleep 3600']
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	return nil
}

func (l *PodTopologyConstraintMinDomainsLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "deployment", "distributed-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if strings.TrimSpace(output) == "" || strings.TrimSpace(output) == "0" {
		return nil
	}
	return nil
}

func (l *PodTopologyConstraintMinDomainsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "distributed-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	if strings.TrimSpace(output) != "6" {
		return fmt.Errorf("deployment not ready (ready replicas: %s, expected: 6)", output)
	}

	return nil
}

func (l *PodTopologyConstraintMinDomainsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check topology constraints",
			Command:     "kubectl get deploy distributed-app -o yaml | grep -A 10 topologySpreadConstraints",
			Notes:       "minDomains is 5 but cluster only has 3 nodes",
		},
		{
			Description: "Count cluster nodes",
			Command:     "kubectl get nodes --no-headers | wc -l",
			Notes:       "Cluster has fewer than 5 nodes",
		},
		{
			Description: "Fix minDomains",
			Command:     "kubectl patch deploy distributed-app --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/template/spec/topologySpreadConstraints/0/minDomains\",\"value\":1}]'",
			Notes:       "Set minDomains to 1 to allow scheduling on any node",
		},
		{
			Description: "Verify deployment is ready",
			Command:     "kubectl rollout status deploy/distributed-app --timeout=120s",
			Notes:       "All 6 replicas should be ready",
		},
	}
}
