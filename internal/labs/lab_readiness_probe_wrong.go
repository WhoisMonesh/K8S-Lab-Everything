package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ReadinessProbeWrongLab{})
}

type ReadinessProbeWrongLab struct {
	BaseLab
}

func (l *ReadinessProbeWrongLab) ID() string {
	return "readiness_probe_wrong"
}

func (l *ReadinessProbeWrongLab) Title() string {
	return "Pod Not Ready Due to Readiness Probe"
}

func (l *ReadinessProbeWrongLab) Category() Category {
	return CategoryWorkloads
}

func (l *ReadinessProbeWrongLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ReadinessProbeWrongLab) Description() string {
	return `A deployment 'api' has pods running but none are Ready. The Service has no endpoints because the readiness probe fails.

Your task: Fix the readiness probe so pods become Ready and receive traffic.`
}

func (l *ReadinessProbeWrongLab) Hints() []string {
	return []string{
		"Check the pod readiness status",
		"Look at the readiness probe configuration",
		"The probe path might be wrong",
		"Check if the application exposes the expected endpoint",
	}
}

func (l *ReadinessProbeWrongLab) EstimatedTime() int {
	return 15
}

func (l *ReadinessProbeWrongLab) Tags() []string {
	return []string{"readiness-probe", "health-check", "endpoints", "service"}
}

func (l *ReadinessProbeWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ReadinessProbeWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create deployment with wrong readiness probe path
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      containers:
      - name: web
        image: nginx:alpine
        ports:
        - containerPort: 80
        readinessProbe:
          httpGet:
            path: /healthz
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	// Create service
	svc := `apiVersion: v1
kind: Service
metadata:
  name: api-svc
  namespace: default
spec:
  selector:
    app: api
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, svc); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}

func (l *ReadinessProbeWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ReadinessProbeWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if pods are Ready
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=api",
		"-o", "jsonpath={.items[*].status.conditions[?(@.type=='Ready')].status}")
	if err != nil {
		return fmt.Errorf("failed to check pods: %w", err)
	}

	for _, status := range strings.Fields(output) {
		if status != "True" {
			return fmt.Errorf("not all pods are Ready (statuses: %s)", output)
		}
	}

	// Check if service has endpoints
	output, err = kubectl(ctx, kubeconfigPath, "get", "endpoints", "api-svc",
		"-o", "jsonpath={.subsets[*].addresses[*].ip}")
	if err != nil {
		return fmt.Errorf("failed to check endpoints: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("service has no endpoints")
	}

	return nil
}

func (l *ReadinessProbeWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check service endpoints",
			Command:     "kubectl get endpoints api-svc",
			Notes:       "The endpoints list will be empty",
		},
		{
			Description: "Check pod readiness",
			Command:     "kubectl get pods -l app=api -o wide",
			Notes:       "Pods are Running but not Ready (0/1 Ready)",
		},
		{
			Description: "Check readiness probe configuration",
			Command:     "kubectl get deployment api -o yaml | grep -A 5 readinessProbe",
			Notes:       "The probe targets /healthz but nginx doesn't have that path",
		},
		{
			Description: "Test the health endpoint",
			Command:     "kubectl exec deploy/api -- curl -s http://localhost/healthz",
			Notes:       "This will return 404, confirming the path doesn't exist",
		},
		{
			Description: "Fix the readiness probe path",
			Command:     "kubectl edit deployment api",
			Notes:       "Change path from /healthz to / (nginx's default path)",
		},
		{
			Description: "Verify pods are Ready",
			Command:     "kubectl get pods -l app=api",
			Notes:       "Pods should now show 1/1 Ready",
		},
		{
			Description: "Verify service has endpoints",
			Command:     "kubectl get endpoints api-svc",
			Notes:       "The endpoints should now show pod IPs",
		},
	}
}
