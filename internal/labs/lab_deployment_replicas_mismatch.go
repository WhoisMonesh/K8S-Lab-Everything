package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&DeploymentReplicasMismatchLab{})
}

type DeploymentReplicasMismatchLab struct {
	BaseLab
}

func (l *DeploymentReplicasMismatchLab) ID() string {
	return "deployment_replicas_mismatch"
}

func (l *DeploymentReplicasMismatchLab) Title() string {
	return "Deployment Desired vs Actual Replicas Mismatch"
}

func (l *DeploymentReplicasMismatchLab) Category() Category {
	return CategoryWorkloads
}

func (l *DeploymentReplicasMismatchLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *DeploymentReplicasMismatchLab) Description() string {
	return `A deployment 'worker' shows 3 desired replicas but only 1 is available.
The pods are running but not becoming Ready due to a misconfigured readiness probe.

Your task: Fix the configuration so all replicas become Ready.`
}

func (l *DeploymentReplicasMismatchLab) Hints() []string {
	return []string{
		"Check the deployment status",
		"Look at the available vs ready replicas",
		"Check pod readiness conditions",
		"The readiness probe might be configured incorrectly",
	}
}

func (l *DeploymentReplicasMismatchLab) EstimatedTime() int {
	return 20
}

func (l *DeploymentReplicasMismatchLab) Tags() []string {
	return []string{"deployment", "replicas", "readiness", "workloads"}
}

func (l *DeploymentReplicasMismatchLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *DeploymentReplicasMismatchLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create deployment with wrong readiness probe
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: worker
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: worker
  template:
    metadata:
      labels:
        app: worker
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80
        readinessProbe:
          httpGet:
            path: /ready
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	return nil
}

func (l *DeploymentReplicasMismatchLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *DeploymentReplicasMismatchLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if deployment has all replicas ready
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "worker",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	ready := strings.TrimSpace(output)
	if ready == "" || ready == "0" {
		return fmt.Errorf("no ready replicas")
	}

	// Check readyReplicas equals replicas
	output, err = kubectl(ctx, kubeconfigPath, "get", "deployment", "worker",
		"-o", "jsonpath={.spec.replicas}")
	if err != nil {
		return fmt.Errorf("failed to check desired replicas: %w", err)
	}

	desired := strings.TrimSpace(output)
	if ready != desired {
		return fmt.Errorf("replicas mismatch (ready: %s, desired: %s)", ready, desired)
	}

	return nil
}

func (l *DeploymentReplicasMismatchLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check deployment status",
			Command:     "kubectl get deployment worker",
			Notes:       "Shows 3 desired but only 1 available",
		},
		{
			Description: "Check pod status",
			Command:     "kubectl get pods -l app=worker",
			Notes:       "All pods are Running but not Ready",
		},
		{
			Description: "Check readiness probe",
			Command:     "kubectl get deployment worker -o yaml | grep -A 5 readinessProbe",
			Notes:       "The probe targets /ready which doesn't exist",
		},
		{
			Description: "Test the readiness endpoint",
			Command:     "kubectl exec deploy/worker -- curl -s http://localhost/ready",
			Notes:       "This will return 404",
		},
		{
			Description: "Fix the readiness probe",
			Command:     "kubectl edit deployment worker",
			Notes:       "Change path from /ready to / (nginx's default)",
		},
		{
			Description: "Wait for rollout",
			Command:     "kubectl rollout status deployment worker",
			Notes:       "Wait for all replicas to become ready",
		},
		{
			Description: "Verify all replicas are ready",
			Command:     "kubectl get deployment worker",
			Notes:       "Should show 3/3 ready replicas",
		},
	}
}
