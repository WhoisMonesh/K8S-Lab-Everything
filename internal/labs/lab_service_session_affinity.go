package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ServiceSessionAffinityLab{})
}

type ServiceSessionAffinityLab struct {
	BaseLab
}

func (l *ServiceSessionAffinityLab) ID() string {
	return "service_session_affinity"
}

func (l *ServiceSessionAffinityLab) Title() string {
	return "Session Affinity Misconfigured"
}

func (l *ServiceSessionAffinityLab) Category() Category {
	return CategoryNetworking
}

func (l *ServiceSessionAffinityLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ServiceSessionAffinityLab) Description() string {
	return `A Service 'sticky-app' has sessionAffinity set to ClientIP but
the sessionAffinityConfig is missing or misconfigured. This causes
inconsistent session handling.

Your task: Fix the session affinity configuration.`
}

func (l *ServiceSessionAffinityLab) Hints() []string {
	return []string{
		"Check the Service session affinity settings",
		"sessionAffinity: ClientIP requires sessionAffinityConfig",
		"Configure timeoutSeconds for session persistence",
	}
}

func (l *ServiceSessionAffinityLab) EstimatedTime() int {
	return 10
}

func (l *ServiceSessionAffinityLab) Tags() []string {
	return []string{"service", "session-affinity", "networking"}
}

func (l *ServiceSessionAffinityLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceSessionAffinityLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: sticky-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: sticky-app
  template:
    metadata:
      labels:
        app: sticky-app
    spec:
      containers:
      - name: web
        image: nginx:alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	service := `apiVersion: v1
kind: Service
metadata:
  name: sticky-app
  namespace: default
spec:
  selector:
    app: sticky-app
  ports:
  - port: 80
    targetPort: 80
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 0
`
	if err := kubectlApply(ctx, kubeconfigPath, service); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}

func (l *ServiceSessionAffinityLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ServiceSessionAffinityLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "sticky-app",
		"-o", "jsonpath={.spec.sessionAffinityConfig.clientIP.timeoutSeconds}")
	if err != nil {
		return fmt.Errorf("failed to check service: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "0" || val == "" {
		return fmt.Errorf("timeoutSeconds is 0 or empty")
	}

	return nil
}

func (l *ServiceSessionAffinityLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check Service session affinity",
			Command:     "kubectl get svc sticky-app -o yaml | grep -A 5 sessionAffinity",
			Notes:       "timeoutSeconds is 0 which disables session persistence",
		},
		{
			Description: "Fix session affinity config",
			Command:     "kubectl edit svc sticky-app",
			Notes:       "Set timeoutSeconds to 10800 (3 hours) or another appropriate value",
		},
		{
			Description: "Verify configuration",
			Command:     "kubectl get svc sticky-app -o yaml | grep -A 5 sessionAffinity",
			Notes:       "timeoutSeconds should now be a positive value",
		},
	}
}
