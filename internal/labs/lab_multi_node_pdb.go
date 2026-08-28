package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&MultiNodePDBLab{})
}

type MultiNodePDBLab struct {
	BaseLab
}

func (l *MultiNodePDBLab) ID() string {
	return "multi_node_pdb"
}

func (l *MultiNodePDBLab) Title() string {
	return "Pod Disruption Budget with Node Drain"
}

func (l *MultiNodePDBLab) Category() Category {
	return CategoryClusterArchitecture
}

func (l *MultiNodePDBLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *MultiNodePDBLab) Description() string {
	return `A critical application runs across multiple worker nodes. A Pod
Disruption Budget should protect it during node maintenance, but the PDB
configuration is wrong and allows all replicas to be evicted at once.

Your task: Fix the PodDisruptionBudget so it properly protects the application
during node drains.`
}

func (l *MultiNodePDBLab) Hints() []string {
	return []string{
		"Check the current PDB configuration",
		"maxUnavailable: 100% allows all pods to be evicted",
		"Use minAvailable or a lower maxUnavailable value",
	}
}

func (l *MultiNodePDBLab) EstimatedTime() int {
	return 25
}

func (l *MultiNodePDBLab) Tags() []string {
	return []string{"pdb", "disruption-budget", "multi-node", "maintenance", "high-availability"}
}

func (l *MultiNodePDBLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: critical-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: critical-app
  template:
    metadata:
      labels:
        app: critical-app
    spec:
      topologySpreadConstraints:
      - maxSkew: 1
        topologyKey: kubernetes.io/hostname
        whenUnsatisfiable: DoNotSchedule
        labelSelector:
          matchLabels:
            app: critical-app
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo critical; sleep 15; done']
        resources:
          limits:
            memory: 32Mi
            cpu: 50m
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *MultiNodePDBLab) Break(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	pdb := `apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: critical-app-pdb
  namespace: default
spec:
  maxUnavailable: 100%
  selector:
    matchLabels:
      app: critical-app
`
	return kubectlApply(ctx, kubeconfigPath, pdb)
}

func (l *MultiNodePDBLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pdb", "critical-app-pdb",
		"-o", "jsonpath={.spec.maxUnavailable}")
	if err != nil {
		return fmt.Errorf("checking pdb: %w", err)
	}

	if strings.TrimSpace(output) == "100%" || strings.TrimSpace(output) == "" {
		return nil
	}

	return fmt.Errorf("PDB maxUnavailable is not 100%% (broken state expected)")
}

func (l *MultiNodePDBLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pdb", "critical-app-pdb",
		"-o", "jsonpath={.spec.maxUnavailable}{.spec.minAvailable}")
	if err != nil {
		return fmt.Errorf("checking pdb: %w", err)
	}

	config := strings.TrimSpace(output)

	if strings.Contains(config, "100%") {
		return fmt.Errorf("PDB still allows 100%% disruption")
	}

	maxUnavailable, _ := kubectl(ctx, kubeconfigPath, "get", "pdb", "critical-app-pdb",
		"-o", "jsonpath={.spec.maxUnavailable}")
	minAvailable, _ := kubectl(ctx, kubeconfigPath, "get", "pdb", "critical-app-pdb",
		"-o", "jsonpath={.spec.minAvailable}")

	if strings.TrimSpace(maxUnavailable) == "" && strings.TrimSpace(minAvailable) == "" {
		return fmt.Errorf("PDB has neither maxUnavailable nor minAvailable set")
	}

	return nil
}

func (l *MultiNodePDBLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check current PDB",
			Command:     "kubectl get pdb critical-app-pdb -o yaml",
			Notes:       "Notice maxUnavailable: 100%",
		},
		{
			Description: "Fix: Set minAvailable instead",
			Command:     "kubectl patch pdb critical-app-pdb --type merge -p '{\"spec\":{\"minAvailable\":2,\"maxUnavailable\":null}}'",
			Notes:       "Keep at least 2 pods running during disruptions",
		},
		{
			Description: "Verify PDB configuration",
			Command:     "kubectl get pdb critical-app-pdb",
			Notes:       "Should show minAvailable: 2",
		},
		{
			Description: "Test: Try draining a node (optional)",
			Command:     "kubectl drain <node> --ignore-daemonsets --delete-emptydir-data",
			Notes:       "PDB should prevent eviction if it would violate the budget",
		},
	}
}
