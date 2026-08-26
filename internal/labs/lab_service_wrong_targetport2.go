package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ServiceWrongTargetPort{})
}

type ServiceWrongTargetPort struct {
	BaseLab
}

func (l *ServiceWrongTargetPort) ID() string            { return "service_wrong_targetport2" }
func (l *ServiceWrongTargetPort) Title() string         { return "Service Points to Wrong targetPort" }
func (l *ServiceWrongTargetPort) Category() Category    { return CategoryNetworking }
func (l *ServiceWrongTargetPort) Difficulty() Difficulty { return DifficultyEasy }
func (l *ServiceWrongTargetPort) EstimatedTime() int    { return 10 }
func (l *ServiceWrongTargetPort) Tags() []string        { return []string{"networking", "service", "targetport"} }

func (l *ServiceWrongTargetPort) Description() string {
	return `A service is pointing to the wrong targetPort.
Traffic is being dropped because the backend port doesn't match.
Fix the service targetPort.`
}

func (l *ServiceWrongTargetPort) Hints() []string {
	return []string{
		"Check the service targetPort",
		"Verify the container port",
		"Update targetPort to match container port",
	}
}

func (l *ServiceWrongTargetPort) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceWrongTargetPort) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 1
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
---
apiVersion: v1
kind: Service
metadata:
  name: web-svc
spec:
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 8080`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ServiceWrongTargetPort) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "web-svc",
		"-o", "jsonpath={.spec.ports[0].targetPort}")
	if err != nil {
		return err
	}
	if output == "8080" {
		return fmt.Errorf("targetPort still wrong")
	}
	return nil
}

func (l *ServiceWrongTargetPort) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check service", Command: "kubectl get svc web-svc -o yaml"},
		{Description: "Fix targetPort", Command: "kubectl patch svc web-svc -p '{\"spec\":{\"ports\":[{\"port\":80,\"targetPort\":80}]}}'"},
	}
}
