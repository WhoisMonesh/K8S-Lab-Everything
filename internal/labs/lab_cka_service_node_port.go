package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&ServiceNodePortLab{})
}

type ServiceNodePortLab struct {
	BaseLab
}

func (l *ServiceNodePortLab) ID() string             { return "cka_service_node_port" }
func (l *ServiceNodePortLab) Title() string          { return "Configure NodePort Service Range" }
func (l *ServiceNodePortLab) Category() Category     { return CategoryServicesNetworking }
func (l *ServiceNodePortLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *ServiceNodePortLab) EstimatedTime() int     { return 15 }
func (l *ServiceNodePortLab) Tags() []string {
	return []string{"service", "nodeport", "networking"}
}
func (l *ServiceNodePortLab) Cert() Cert        { return CertCKA }
func (l *ServiceNodePortLab) DomainWeight() int { return 20 }

func (l *ServiceNodePortLab) Description() string {
	return `A NodePort service is configured with a port outside the valid range.
Fix the service to use a valid NodePort within the 30000-32767 range.`
}

func (l *ServiceNodePortLab) Hints() []string {
	return []string{
		"Check the NodePort range in API server config",
		"Valid range is 30000-32767",
		"Patch the service to use a valid port",
	}
}

func (l *ServiceNodePortLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceNodePortLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: nodeport-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: nodeport-ns
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
		return fmt.Errorf("creating deployment: %w", err)
	}

	svc := `apiVersion: v1
kind: Service
metadata:
  name: nodeport-svc
  namespace: nodeport-ns
spec:
  type: ClusterIP
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, svc); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}

func (l *ServiceNodePortLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ServiceNodePortLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "nodeport-svc",
		"-n", "nodeport-ns", "-o", "jsonpath={.spec.ports[0].nodePort}")
	if err != nil {
		return err
	}
	port := 0
	fmt.Sscanf(output, "%d", &port)
	if port < 30000 || port > 32767 {
		return fmt.Errorf("NodePort %d is outside valid range 30000-32767", port)
	}
	return nil
}

func (l *ServiceNodePortLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check service", Command: "kubectl get svc nodeport-svc -n nodeport-ns -o yaml"},
		{Description: "Fix NodePort", Command: "kubectl patch svc nodeport-svc -n nodeport-ns -p '{\"spec\":{\"ports\":[{\"port\":80,\"targetPort\":80,\"nodePort\":30080}]}}'"},
		{Description: "Verify", Command: "kubectl get svc nodeport-svc -n nodeport-ns"},
	}
}
