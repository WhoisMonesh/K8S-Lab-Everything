package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&IngressBrokenLab{})
}

type IngressBrokenLab struct {
	BaseLab
}

func (l *IngressBrokenLab) ID() string {
	return "ingress_broken"
}

func (l *IngressBrokenLab) Title() string {
	return "Ingress Configuration Issues"
}

func (l *IngressBrokenLab) Category() Category {
	return CategoryNetworking
}

func (l *IngressBrokenLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *IngressBrokenLab) Description() string {
	return `An Ingress resource has been created to expose a web application,
but traffic is not routing correctly. The application is running, but
requests to the ingress endpoint are failing.

Your task: Fix the Ingress configuration so that traffic routes correctly
to the backend service.`
}

func (l *IngressBrokenLab) Hints() []string {
	return []string{
		"Check the Ingress resource configuration",
		"Verify the backend service name and port",
		"Look at the service selector and ensure it matches the pods",
		"Check if the ingress class is correctly specified",
	}
}

func (l *IngressBrokenLab) EstimatedTime() int {
	return 20
}

func (l *IngressBrokenLab) Tags() []string {
	return []string{"ingress", "networking", "services", "routing", "troubleshooting"}
}

func (l *IngressBrokenLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	// Create a namespace for the lab
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: ingress-lab
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return err
	}

	// Create a working deployment
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: ingress-lab
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return err
	}

	// Create a service (this will be working)
	service := `apiVersion: v1
kind: Service
metadata:
  name: web-service
  namespace: ingress-lab
spec:
  selector:
    app: web
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
`
	return kubectlApply(ctx, kubeconfigPath, service)
}

func (l *IngressBrokenLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create a broken Ingress with wrong service name
	brokenIngress := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: web-ingress
  namespace: ingress-lab
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: web.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: web-svc
            port:
              number: 80
`
	return kubectlApply(ctx, kubeconfigPath, brokenIngress)
}

func (l *IngressBrokenLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Check if the ingress exists with wrong service name
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress",
		"web-ingress", "-n", "ingress-lab",
		"-o", "jsonpath={.spec.rules[0].http.paths[0].backend.service.name}")
	if err != nil {
		return fmt.Errorf("ingress not found: %w", err)
	}

	if strings.TrimSpace(output) != "web-svc" {
		return fmt.Errorf("ingress service name is not broken (expected: web-svc)")
	}

	return nil
}

func (l *IngressBrokenLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if the ingress has been fixed with correct service name
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress",
		"web-ingress", "-n", "ingress-lab",
		"-o", "jsonpath={.spec.rules[0].http.paths[0].backend.service.name}")
	if err != nil {
		return fmt.Errorf("ingress not found: %w", err)
	}

	serviceName := strings.TrimSpace(output)
	if serviceName != "web-service" {
		return fmt.Errorf("ingress still points to wrong service: %s (expected: web-service)", serviceName)
	}

	// Verify the service exists
	_, err = kubectl(ctx, kubeconfigPath, "get", "service",
		"web-service", "-n", "ingress-lab")
	if err != nil {
		return fmt.Errorf("backend service not found: %w", err)
	}

	return nil
}

func (l *IngressBrokenLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check the Ingress resource",
			Command:     "kubectl get ingress -n ingress-lab",
			Notes:       "List all ingress resources in the namespace",
		},
		{
			Description: "Describe the Ingress to see details",
			Command:     "kubectl describe ingress web-ingress -n ingress-lab",
			Notes:       "Look for the backend service configuration and any warnings",
		},
		{
			Description: "List services in the namespace",
			Command:     "kubectl get svc -n ingress-lab",
			Notes:       "Find the actual service name that the ingress should point to",
		},
		{
			Description: "View the Ingress YAML configuration",
			Command:     "kubectl get ingress web-ingress -n ingress-lab -o yaml",
			Notes:       "Look at the service name in the backend configuration",
		},
		{
			Description: "Identify the mismatch",
			Notes:       "The Ingress points to 'web-svc' but the actual service is 'web-service'",
		},
		{
			Description: "Fix the Ingress by editing it",
			Command:     "kubectl edit ingress web-ingress -n ingress-lab",
			Notes:       "Change the service name from 'web-svc' to 'web-service'",
		},
		{
			Description: "Verify the fix",
			Command:     "kubectl describe ingress web-ingress -n ingress-lab",
			Notes:       "The backend should now point to the correct service",
		},
		{
			Description: "Check the backend pods are running",
			Command:     "kubectl get pods -n ingress-lab -l app=web",
			Notes:       "Ensure the application pods are running",
		},
	}
}
