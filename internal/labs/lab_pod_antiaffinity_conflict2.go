package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodAntiAffinityConflict{})
}

type PodAntiAffinityConflict struct {
	BaseLab
}

func (l *PodAntiAffinityConflict) ID() string { return "pod_antiaffinity_conflict2" }
func (l *PodAntiAffinityConflict) Title() string {
	return "Deployment Anti-Affinity Scheduling Failure"
}
func (l *PodAntiAffinityConflict) Category() Category     { return CategoryScheduling }
func (l *PodAntiAffinityConflict) Difficulty() Difficulty { return DifficultyHard }
func (l *PodAntiAffinityConflict) EstimatedTime() int     { return 20 }
func (l *PodAntiAffinityConflict) Tags() []string {
	return []string{"scheduling", "antiaffinity", "pods"}
}

func (l *PodAntiAffinityConflict) Description() string {
	return `A deployment cannot schedule pods due to conflicting pod anti-affinity rules.
Fix the anti-affinity configuration to allow scheduling.`
}

func (l *PodAntiAffinityConflict) Hints() []string {
	return []string{
		"Check pod anti-affinity rules",
		"Look at requiredDuringSchedulingIgnoredDuringExecution",
		"Consider using preferredDuringSchedulingIgnoredDuringExecution",
	}
}

func (l *PodAntiAffinityConflict) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodAntiAffinityConflict) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: antiaffinity-app
spec:
  replicas: 5
  selector:
    matchLabels:
      app: antiaffinity-app
  template:
    metadata:
      labels:
        app: antiaffinity-app
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - antiaffinity-app
            topologyKey: kubernetes.io/hostname
      containers:
      - name: nginx
        image: nginx:alpine`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PodAntiAffinityConflict) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/antiaffinity-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return err
	}
	if output == "" || output == "0" {
		return fmt.Errorf("deployment not ready")
	}
	return nil
}

func (l *PodAntiAffinityConflict) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check anti-affinity", Command: "kubectl get deploy antiaffinity-app -o jsonpath='{.spec.template.spec.affinity}'"},
		{Description: "Fix anti-affinity", Command: "kubectl edit deploy antiaffinity-app"},
		{Description: "Change to preferred", Command: "Change requiredDuringSchedulingIgnoredDuringExecution to preferredDuringSchedulingIgnoredDuringExecution"},
	}
}
