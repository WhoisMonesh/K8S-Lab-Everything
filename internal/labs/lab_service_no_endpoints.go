package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ServiceNoEndpointsLab{})
}

type ServiceNoEndpointsLab struct {
	BaseLab
}

func (l *ServiceNoEndpointsLab) ID() string {
	return "service_no_endpoints"
}

func (l *ServiceNoEndpointsLab) Title() string {
	return "Service With No Endpoints"
}

func (l *ServiceNoEndpointsLab) Category() Category {
	return CategoryNetworking
}

func (l *ServiceNoEndpointsLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ServiceNoEndpointsLab) Description() string {
	return `A backend application is deployed and running in the 'web' namespace,
but the frontend gets connection refused when calling http://backend-svc:80.

kubectl get endpoints backend-svc -n web shows <none>.

Your task: Debug why the Service has no endpoints and fix it so traffic reaches the pods.`
}

func (l *ServiceNoEndpointsLab) Hints() []string {
	return []string{
		"Check that the pods behind the service are actually running",
		"Compare the Service selector with the pod labels",
		"kubectl get endpoints backend-svc -n web tells you whether any pods are selected",
		"A selector key/value must match the pod labels exactly",
	}
}

func (l *ServiceNoEndpointsLab) EstimatedTime() int {
	return 20
}

func (l *ServiceNoEndpointsLab) Tags() []string {
	return []string{"service", "endpoints", "selector", "networking", "troubleshooting"}
}

func (l *ServiceNoEndpointsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceNoEndpointsLab) Break(ctx context.Context, kubeconfigPath string) error {
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: web
`

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: web
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
      - name: api
        image: nginx:alpine
        ports:
        - containerPort: 80
`

	service := `apiVersion: v1
kind: Service
metadata:
  name: backend-svc
  namespace: web
spec:
  selector:
    app: backends
  ports:
  - port: 80
    targetPort: 80
`

	for _, manifest := range []string{namespace, deployment, service} {
		if err := kubectlApply(ctx, kubeconfigPath, manifest); err != nil {
			return fmt.Errorf("applying lab resources: %w", err)
		}
	}
	return nil
}

func (l *ServiceNoEndpointsLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, err := kubectl(ctx, kubeconfigPath, "get", "endpoints", "backend-svc", "-n", "web")
	if err != nil {
		return nil
	}
	if strings.Contains(output, "<none>") || !strings.Contains(output, ":") {
		return nil
	}
	return fmt.Errorf("expected service to have no endpoints")
}

func (l *ServiceNoEndpointsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "backend",
		"-n", "web", "-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}
	if output != "2" {
		return fmt.Errorf("deployment not fully ready yet (ready replicas: %s, expected: 2)", output)
	}

	endpoints, err := kubectl(ctx, kubeconfigPath, "get", "endpoints", "backend-svc",
		"-n", "web", "-o", "jsonpath={.subsets[*].addresses[*].ip}")
	if err != nil {
		return fmt.Errorf("failed to check endpoints: %w", err)
	}
	endpoints = strings.TrimSpace(endpoints)
	if endpoints == "" {
		return fmt.Errorf("service backend-svc still has no endpoints - check the selector")
	}
	return nil
}

func (l *ServiceNoEndpointsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Confirm the pods are running",
			Command:     "kubectl get pods -n web",
			Notes:       "Both backend pods should be Running - the problem is not the pods themselves",
		},
		{
			Description: "Check the service endpoints",
			Command:     "kubectl get endpoints backend-svc -n web",
			Notes:       "You will see <none> which means the selector matches zero pods",
		},
		{
			Description: "Compare the selector against the pod labels",
			Command:     "kubectl get svc backend-svc -n web -o jsonpath='{.spec.selector}' && kubectl get pods -n web --show-labels",
			Notes:       "The service selects app=backends but the pods carry app=backend",
		},
		{
			Description: "Fix the service selector",
			Command:     "kubectl patch svc backend-svc -n web -p '{\"spec\":{\"selector\":{\"app\":\"backend\"}}}'",
			Notes:       "Alternatively edit the service with kubectl edit svc backend-svc -n web",
		},
		{
			Description: "Confirm endpoints now list pod IPs",
			Command:     "kubectl get endpoints backend-svc -n web",
			Notes:       "You should see two IP addresses, one per replica",
		},
	}
}
