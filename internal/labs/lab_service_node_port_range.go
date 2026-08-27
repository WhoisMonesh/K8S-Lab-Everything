package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ServiceNodePortRangeLab{})
}

type ServiceNodePortRangeLab struct {
	BaseLab
}

func (l *ServiceNodePortRangeLab) ID() string {
	return "service_node_port_range"
}

func (l *ServiceNodePortRangeLab) Title() string {
	return "NodePort Out of Range"
}

func (l *ServiceNodePortRangeLab) Category() Category {
	return CategoryNetworking
}

func (l *ServiceNodePortRangeLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *ServiceNodePortRangeLab) Description() string {
	return `A Service 'node-app' with type NodePort has a nodePort set to 80,
which is outside the valid NodePort range (30000-32767). The Service
cannot be created.

Your task: Fix the nodePort to use a valid port number.`
}

func (l *ServiceNodePortRangeLab) Hints() []string {
	return []string{
		"Check the Service configuration",
		"NodePort range is typically 30000-32767",
		"Change nodePort to a value within the valid range",
	}
}

func (l *ServiceNodePortRangeLab) EstimatedTime() int {
	return 5
}

func (l *ServiceNodePortRangeLab) Tags() []string {
	return []string{"service", "nodeport", "networking"}
}

func (l *ServiceNodePortRangeLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceNodePortRangeLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: node-app
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: node-app
  template:
    metadata:
      labels:
        app: node-app
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
  name: node-app
  namespace: default
spec:
  type: NodePort
  selector:
    app: node-app
  ports:
  - port: 80
    targetPort: 80
    nodePort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, service); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}

func (l *ServiceNodePortRangeLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *ServiceNodePortRangeLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "node-app",
		"-o", "jsonpath={.spec.ports[0].nodePort}")
	if err != nil {
		return fmt.Errorf("failed to check service: %w", err)
	}

	val := strings.TrimSpace(output)
	if val == "80" {
		return fmt.Errorf("nodePort is still 80")
	}

	return nil
}

func (l *ServiceNodePortRangeLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check Service configuration",
			Command:     "kubectl get service node-app -o yaml | grep nodePort",
			Notes:       "nodePort is 80 which is outside valid range",
		},
		{
			Description: "Fix nodePort",
			Command:     "kubectl patch service node-app --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/ports/0/nodePort\",\"value\":30080}]'",
			Notes:       "Set to 30080 which is within 30000-32767 range",
		},
		{
			Description: "Verify Service",
			Command:     "kubectl get service node-app",
			Notes:       "Should show nodePort as 30080",
		},
	}
}
