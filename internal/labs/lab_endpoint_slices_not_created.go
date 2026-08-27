package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&EndpointSlicesNotCreatedLab{})
}

type EndpointSlicesNotCreatedLab struct {
	BaseLab
}

func (l *EndpointSlicesNotCreatedLab) ID() string {
	return "endpoint_slices_not_created"
}

func (l *EndpointSlicesNotCreatedLab) Title() string {
	return "EndpointSlice Not Created"
}

func (l *EndpointSlicesNotCreatedLab) Category() Category {
	return CategoryNetworking
}

func (l *EndpointSlicesNotCreatedLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *EndpointSlicesNotCreatedLab) Description() string {
	return `A Service 'backend-svc' has no EndpointSlice created because the
Service selector doesn't match any pod labels. The pods have label
app=backend but the Service selector uses app=fronted (typo).

Your task: Fix the Service selector to create the EndpointSlice.`
}

func (l *EndpointSlicesNotCreatedLab) Hints() []string {
	return []string{
		"Check the Service selector",
		"Compare Service selector with pod labels",
		"EndpointSlices are created automatically when selectors match pods",
	}
}

func (l *EndpointSlicesNotCreatedLab) EstimatedTime() int {
	return 15
}

func (l *EndpointSlicesNotCreatedLab) Tags() []string {
	return []string{"service", "endpointslice", "selector", "networking"}
}

func (l *EndpointSlicesNotCreatedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *EndpointSlicesNotCreatedLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
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
  name: backend-svc
  namespace: default
spec:
  selector:
    app: fronted
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, service); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}

func (l *EndpointSlicesNotCreatedLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "endpointslices", "-l",
		"kubernetes.io/service-name=backend-svc", "-o", "name")
	if strings.Contains(output, "backend-svc") {
		return nil
	}
	return nil
}

func (l *EndpointSlicesNotCreatedLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "endpointslices", "-l",
		"kubernetes.io/service-name=backend-svc", "-o", "jsonpath={.items[0].endpoints[*].addresses[*]}")
	if err != nil {
		return fmt.Errorf("failed to check endpointslices: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no endpoints found in EndpointSlice")
	}

	return nil
}

func (l *EndpointSlicesNotCreatedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check Service selector",
			Command:     "kubectl get svc backend-svc -o yaml | grep -A 3 selector",
			Notes:       "Selector is app=fronted (typo)",
		},
		{
			Description: "Check pod labels",
			Command:     "kubectl get pods --show-labels | grep backend",
			Notes:       "Pods have app=backend, not app=fronted",
		},
		{
			Description: "Check EndpointSlice",
			Command:     "kubectl get endpointslices -l kubernetes.io/service-name=backend-svc",
			Notes:       "No EndpointSlice should exist due to selector mismatch",
		},
		{
			Description: "Fix Service selector",
			Command:     "kubectl edit svc backend-svc",
			Notes:       "Change app=fronted to app=backend",
		},
		{
			Description: "Verify EndpointSlice is created",
			Command:     "kubectl get endpointslices -l kubernetes.io/service-name=backend-svc",
			Notes:       "EndpointSlice should now exist with backend IPs",
		},
	}
}
