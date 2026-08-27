package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ServiceClusterIPConflictLab{})
}

type ServiceClusterIPConflictLab struct {
	BaseLab
}

func (l *ServiceClusterIPConflictLab) ID() string {
	return "service_cluster_ip_conflict"
}

func (l *ServiceClusterIPConflictLab) Title() string {
	return "ClusterIP Address Conflict"
}

func (l *ServiceClusterIPConflictLab) Category() Category {
	return CategoryNetworking
}

func (l *ServiceClusterIPConflictLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *ServiceClusterIPConflictLab) Description() string {
	return `A Service 'app-svc' cannot be created because its requested ClusterIP
(10.96.0.100) conflicts with an existing Service 'dns-svc' that already
uses that IP address.

Your task: Fix the Service configuration to resolve the IP conflict.`
}

func (l *ServiceClusterIPConflictLab) Hints() []string {
	return []string{
		"Check existing Services and their ClusterIPs",
		"Two Services cannot share the same ClusterIP",
		"Either remove the conflicting Service or let Kubernetes assign a new IP",
	}
}

func (l *ServiceClusterIPConflictLab) EstimatedTime() int {
	return 15
}

func (l *ServiceClusterIPConflictLab) Tags() []string {
	return []string{"service", "clusterip", "ip-conflict", "networking"}
}

func (l *ServiceClusterIPConflictLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceClusterIPConflictLab) Break(ctx context.Context, kubeconfigPath string) error {
	existingSvc := `apiVersion: v1
kind: Service
metadata:
  name: dns-svc
  namespace: default
spec:
  clusterIP: 10.96.0.100
  ports:
  - port: 53
    targetPort: 53
  selector:
    app: dns
`
	if err := kubectlApply(ctx, kubeconfigPath, existingSvc); err != nil {
		return fmt.Errorf("creating existing service: %w", err)
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-svc-deploy
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app-svc
  template:
    metadata:
      labels:
        app: app-svc
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	newSvc := `apiVersion: v1
kind: Service
metadata:
  name: app-svc
  namespace: default
spec:
  clusterIP: 10.96.0.100
  selector:
    app: app-svc
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, newSvc); err != nil {
		return fmt.Errorf("creating conflicting service: %w", err)
	}

	return nil
}

func (l *ServiceClusterIPConflictLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *ServiceClusterIPConflictLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check that app-svc exists and has a valid ClusterIP
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "app-svc",
		"-o", "jsonpath={.spec.clusterIP}")
	if err != nil {
		return fmt.Errorf("failed to check service: %w", err)
	}

	clusterIP := strings.TrimSpace(output)
	if clusterIP == "" || clusterIP == "None" {
		return fmt.Errorf("app-svc has no valid ClusterIP")
	}

	// Verify endpoints exist
	output, err = kubectl(ctx, kubeconfigPath, "get", "endpoints", "app-svc",
		"-o", "jsonpath={.subsets[*].addresses[*].ip}")
	if err != nil {
		return fmt.Errorf("failed to check endpoints: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no endpoints found for app-svc")
	}

	return nil
}

func (l *ServiceClusterIPConflictLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check existing Services",
			Command:     "kubectl get svc -A | grep 10.96.0.100",
			Notes:       "dns-svc already uses the requested ClusterIP",
		},
		{
			Description: "Check app-svc status",
			Command:     "kubectl describe svc app-svc | grep -A 5 Events",
			Notes:       "Should show IP address conflict error",
		},
		{
			Description: "Fix: Remove explicit ClusterIP from app-svc",
			Command:     "kubectl edit svc app-svc",
			Notes:       "Remove the clusterIP field to let Kubernetes assign automatically",
		},
		{
			Description: "Verify service is working",
			Command:     "kubectl get svc app-svc",
			Notes:       "Should have a valid ClusterIP assigned",
		},
	}
}
