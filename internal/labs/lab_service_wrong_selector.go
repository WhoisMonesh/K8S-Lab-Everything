package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ServiceWrongSelectorLab{})
}

type ServiceWrongSelectorLab struct {
	BaseLab
}

func (l *ServiceWrongSelectorLab) ID() string {
	return "service_wrong_selector"
}

func (l *ServiceWrongSelectorLab) Title() string {
	return "Service With Wrong Selector"
}

func (l *ServiceWrongSelectorLab) Category() Category {
	return CategoryNetworking
}

func (l *ServiceWrongSelectorLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *ServiceWrongSelectorLab) Description() string {
	return `A Service 'api-svc' has no endpoints. The pods are running but the Service selector doesn't match the pod labels.

Your task: Fix the Service selector to route traffic to the correct pods.`
}

func (l *ServiceWrongSelectorLab) Hints() []string {
	return []string{
		"Check the Service endpoints",
		"Look at the Service selector",
		"Check the pod labels",
		"The selector key or value doesn't match the pod labels",
	}
}

func (l *ServiceWrongSelectorLab) EstimatedTime() int {
	return 10
}

func (l *ServiceWrongSelectorLab) Tags() []string {
	return []string{"service", "selector", "endpoints", "networking"}
}

func (l *ServiceWrongSelectorLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceWrongSelectorLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create deployment with correct labels
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: api
      tier: backend
  template:
    metadata:
      labels:
        app: api
        tier: backend
    spec:
      containers:
      - name: api
        image: nginx:alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	// Create Service with wrong selector (tier: frontend instead of backend)
	svc := `apiVersion: v1
kind: Service
metadata:
  name: api-svc
  namespace: default
spec:
  selector:
    app: api
    tier: frontend
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, svc); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}

func (l *ServiceWrongSelectorLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ServiceWrongSelectorLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if service has endpoints
	output, err := kubectl(ctx, kubeconfigPath, "get", "endpoints", "api-svc",
		"-o", "jsonpath={.subsets[*].addresses[*].ip}")
	if err != nil {
		return fmt.Errorf("failed to check endpoints: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("service has no endpoints")
	}

	// Check service selector
	output, err = kubectl(ctx, kubeconfigPath, "get", "service", "api-svc",
		"-o", "jsonpath={.spec.selector.tier}")
	if err != nil {
		return fmt.Errorf("failed to check service selector: %w", err)
	}

	if strings.TrimSpace(output) != "backend" {
		return fmt.Errorf("service selector is wrong (tier=%s, expected: backend)", output)
	}

	return nil
}

func (l *ServiceWrongSelectorLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check service endpoints",
			Command:     "kubectl get endpoints api-svc",
			Notes:       "The endpoints list will be empty",
		},
		{
			Description: "Check the service selector",
			Command:     "kubectl get svc api-svc -o yaml | grep -A 5 selector",
			Notes:       "The selector shows tier: frontend",
		},
		{
			Description: "Check the pod labels",
			Command:     "kubectl get pods --show-labels",
			Notes:       "Pods have labels tier=backend, not tier=frontend",
		},
		{
			Description: "Fix the service selector",
			Command:     "kubectl edit svc api-svc",
			Notes:       "Change tier from 'frontend' to 'backend' in the selector",
		},
		{
			Description: "Verify endpoints are populated",
			Command:     "kubectl get endpoints api-svc",
			Notes:       "The endpoints should now show pod IPs",
		},
	}
}
