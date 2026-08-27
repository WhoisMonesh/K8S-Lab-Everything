package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ServiceExternalTrafficPolicyLab{})
}

type ServiceExternalTrafficPolicyLab struct {
	BaseLab
}

func (l *ServiceExternalTrafficPolicyLab) ID() string {
	return "service_external_trafficPolicy"
}

func (l *ServiceExternalTrafficPolicyLab) Title() string {
	return "External Traffic Policy Issue"
}

func (l *ServiceExternalTrafficPolicyLab) Category() Category {
	return CategoryNetworking
}

func (l *ServiceExternalTrafficPolicyLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ServiceExternalTrafficPolicyLab) Description() string {
	return `A Service 'web-external' with type LoadBalancer has externalTrafficPolicy
set to Cluster. This causes source IP addresses to be masqueraded, making
it impossible to track client IPs.

Your task: Fix the externalTrafficPolicy to preserve source IPs.`
}

func (l *ServiceExternalTrafficPolicyLab) Hints() []string {
	return []string{
		"Check the Service configuration",
		"externalTrafficPolicy: Cluster masquerades source IPs",
		"Set externalTrafficPolicy to Local to preserve source IPs",
	}
}

func (l *ServiceExternalTrafficPolicyLab) EstimatedTime() int {
	return 10
}

func (l *ServiceExternalTrafficPolicyLab) Tags() []string {
	return []string{"service", "external-traffic", "loadbalancer", "networking"}
}

func (l *ServiceExternalTrafficPolicyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceExternalTrafficPolicyLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-external
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web-external
  template:
    metadata:
      labels:
        app: web-external
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
  name: web-external
  namespace: default
spec:
  type: LoadBalancer
  externalTrafficPolicy: Cluster
  selector:
    app: web-external
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, service); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}

func (l *ServiceExternalTrafficPolicyLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ServiceExternalTrafficPolicyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "web-external",
		"-o", "jsonpath={.spec.externalTrafficPolicy}")
	if err != nil {
		return fmt.Errorf("failed to check service: %w", err)
	}

	if strings.TrimSpace(output) == "Cluster" {
		return fmt.Errorf("externalTrafficPolicy is still Cluster")
	}

	return nil
}

func (l *ServiceExternalTrafficPolicyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check Service configuration",
			Command:     "kubectl get service web-external -o yaml | grep externalTrafficPolicy",
			Notes:       "externalTrafficPolicy is Cluster",
		},
		{
			Description: "Fix externalTrafficPolicy",
			Command:     "kubectl patch service web-external --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/externalTrafficPolicy\",\"value\":\"Local\"}]'",
			Notes:       "Set to Local to preserve source IPs",
		},
		{
			Description: "Verify configuration",
			Command:     "kubectl get service web-external -o yaml | grep externalTrafficPolicy",
			Notes:       "Should now show Local",
		},
	}
}
