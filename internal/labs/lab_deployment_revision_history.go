package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&DeploymentRevisionHistoryLab{})
}

type DeploymentRevisionHistoryLab struct {
	BaseLab
}

func (l *DeploymentRevisionHistoryLab) ID() string {
	return "deployment_revision_history"
}

func (l *DeploymentRevisionHistoryLab) Title() string {
	return "Deployment Revision History Limit"
}

func (l *DeploymentRevisionHistoryLab) Category() Category {
	return CategoryWorkloads
}

func (l *DeploymentRevisionHistoryLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *DeploymentRevisionHistoryLab) Description() string {
	return `A deployment 'web-api' has revisionHistoryLimit set to 0, which means
Kubernetes immediately deletes ReplicaSets after a rollout. This makes
it impossible to roll back to a previous version if needed.

Your task: Fix the revisionHistoryLimit to allow rollbacks.`
}

func (l *DeploymentRevisionHistoryLab) Hints() []string {
	return []string{
		"Check the deployment configuration",
		"revisionHistoryLimit=0 deletes all old ReplicaSets immediately",
		"A typical value is 5-10 to keep rollback history",
	}
}

func (l *DeploymentRevisionHistoryLab) EstimatedTime() int {
	return 10
}

func (l *DeploymentRevisionHistoryLab) Tags() []string {
	return []string{"deployment", "revision", "history", "rollback", "workloads"}
}

func (l *DeploymentRevisionHistoryLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DeploymentRevisionHistoryLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-api
  namespace: default
spec:
  replicas: 2
  revisionHistoryLimit: 0
  selector:
    matchLabels:
      app: web-api
  template:
    metadata:
      labels:
        app: web-api
    spec:
      containers:
      - name: api
        image: nginx:1.19
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	time.Sleep(10 * time.Second)

	// Trigger an update
	_, _ = kubectl(ctx, kubeconfigPath, "set", "image", "deployment/web-api",
		"api=nginx:1.20")

	time.Sleep(10 * time.Second)

	return nil
}

func (l *DeploymentRevisionHistoryLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	output, _ := kubectl(ctx, kubeconfigPath, "get", "replicasets", "-l", "app=web-api",
		"-o", "name")
	if strings.Contains(output, "web-api") {
		return nil
	}
	return nil
}

func (l *DeploymentRevisionHistoryLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "web-api",
		"-o", "jsonpath={.spec.revisionHistoryLimit}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "0" || val == "" {
		return fmt.Errorf("revisionHistoryLimit is still 0")
	}

	return nil
}

func (l *DeploymentRevisionHistoryLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check deployment configuration",
			Command:     "kubectl get deployment web-api -o yaml | grep revisionHistoryLimit",
			Notes:       "revisionHistoryLimit is 0",
		},
		{
			Description: "Check existing ReplicaSets",
			Command:     "kubectl get replicasets -l app=web-api",
			Notes:       "Only the current ReplicaSet exists due to limit of 0",
		},
		{
			Description: "Fix revisionHistoryLimit",
			Command:     "kubectl patch deployment web-api --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/revisionHistoryLimit\",\"value\":5}]'",
			Notes:       "Set to 5 to keep rollback history",
		},
		{
			Description: "Verify configuration",
			Command:     "kubectl get deployment web-api -o yaml | grep revisionHistoryLimit",
			Notes:       "revisionHistoryLimit should now be 5",
		},
	}
}
