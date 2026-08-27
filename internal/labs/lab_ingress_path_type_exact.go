package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&IngressPathTypeExactLab{})
}

type IngressPathTypeExactLab struct {
	BaseLab
}

func (l *IngressPathTypeExactLab) ID() string {
	return "ingress_path_type_exact"
}

func (l *IngressPathTypeExactLab) Title() string {
	return "Ingress Path Type Mismatch"
}

func (l *IngressPathTypeExactLab) Category() Category {
	return CategoryNetworking
}

func (l *IngressPathTypeExactLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *IngressPathTypeExactLab) Description() string {
	return `An Ingress 'api-routes' uses pathType: Exact for the /api path.
This means only requests to exactly /api will be routed, not /api/anything.
The application needs to handle all paths under /api.

Your task: Fix the pathType to route all /api subpaths.`
}

func (l *IngressPathTypeExactLab) Hints() []string {
	return []string{
		"Check the Ingress path configuration",
		"Exact matches only the exact path, not subpaths",
		"Use Prefix pathType to match /api and all subpaths",
	}
}

func (l *IngressPathTypeExactLab) EstimatedTime() int {
	return 10
}

func (l *IngressPathTypeExactLab) Tags() []string {
	return []string{"ingress", "path", "path-type", "networking"}
}

func (l *IngressPathTypeExactLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *IngressPathTypeExactLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-backend
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: api-backend
  template:
    metadata:
      labels:
        app: api-backend
    spec:
      containers:
      - name: api
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: api-backend
  namespace: default
spec:
  selector:
    app: api-backend
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	ingress := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-routes
  namespace: default
spec:
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /api
        pathType: Exact
        backend:
          service:
            name: api-backend
            port:
              number: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, ingress); err != nil {
		return fmt.Errorf("creating ingress: %w", err)
	}

	return nil
}

func (l *IngressPathTypeExactLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *IngressPathTypeExactLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "api-routes",
		"-o", "jsonpath={.spec.rules[0].http.paths[0].pathType}")
	if err != nil {
		return fmt.Errorf("failed to check ingress: %w", err)
	}

	if strings.TrimSpace(output) == "Exact" {
		return fmt.Errorf("pathType is still Exact")
	}

	return nil
}

func (l *IngressPathTypeExactLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check Ingress path configuration",
			Command:     "kubectl get ingress api-routes -o yaml | grep -A 5 paths",
			Notes:       "pathType is Exact which only matches /api exactly",
		},
		{
			Description: "Fix pathType",
			Command:     "kubectl edit ingress api-routes",
			Notes:       "Change pathType from Exact to Prefix",
		},
		{
			Description: "Verify Ingress",
			Command:     "kubectl get ingress api-routes",
			Notes:       "pathType should now be Prefix",
		},
	}
}
