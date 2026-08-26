package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodTopologySpreadViolation{})
}

type PodTopologySpreadViolation struct {
	BaseLab
}

func (l *PodTopologySpreadViolation) ID() string            { return "pod_topology_spread_violation" }
func (l *PodTopologySpreadViolation) Title() string         { return "Pod Topology Spread Constraint Violation" }
func (l *PodTopologySpreadViolation) Category() Category    { return CategoryScheduling }
func (l *PodTopologySpreadViolation) Difficulty() Difficulty { return DifficultyHard }
func (l *PodTopologySpreadViolation) EstimatedTime() int    { return 20 }
func (l *PodTopologySpreadViolation) Tags() []string        { return []string{"scheduling", "topology", "spread"} }

func (l *PodTopologySpreadViolation) Description() string {
	return `A deployment cannot schedule pods because topology spread constraints cannot be satisfied.
Fix the topology spread constraints to allow scheduling.`
}

func (l *PodTopologySpreadViolation) Hints() []string {
	return []string{
		"Check topology spread constraints",
		"Verify maxSkew value",
		"Consider adjusting the constraint or adding nodes",
	}
}

func (l *PodTopologySpreadViolation) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodTopologySpreadViolation) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: spread-app
spec:
  replicas: 10
  selector:
    matchLabels:
      app: spread-app
  template:
    metadata:
      labels:
        app: spread-app
    spec:
      topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: kubernetes.io/hostname
        whenUnsatisfiable: DoNotSchedule
        labelSelector:
          matchLabels:
            app: spread-app
      containers:
      - name: nginx
        image: nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PodTopologySpreadViolation) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/spread-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return err
	}
	if output == "" || output == "0" {
		return fmt.Errorf("deployment not ready")
	}
	return nil
}

func (l *PodTopologySpreadViolation) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check constraints", Command: "kubectl get deploy spread-app -o jsonpath='{.spec.template.spec.topologySpreadConstraints}'"},
		{Description: "Fix maxSkew", Command: "kubectl edit deploy spread-app"},
		{Description: "Increase maxSkew", Command: "Change maxSkew from 1 to 2 or set whenUnsatisfiable to ScheduleAnyway"},
	}
}
