package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&HeadlessServiceDNS{})
}

type HeadlessServiceDNS struct {
	BaseLab
}

func (l *HeadlessServiceDNS) ID() string            { return "headless_service_dns" }
func (l *HeadlessServiceDNS) Title() string         { return "Headless Service DNS Not Working" }
func (l *HeadlessServiceDNS) Category() Category    { return CategoryDNS }
func (l *HeadlessServiceDNS) Difficulty() Difficulty { return DifficultyMedium }
func (l *HeadlessServiceDNS) EstimatedTime() int    { return 15 }
func (l *HeadlessServiceDNS) Tags() []string        { return []string{"dns", "headless", "statefulset"} }

func (l *HeadlessServiceDNS) Description() string {
	return `A StatefulSet's pods cannot resolve each other via DNS.
The headless service is misconfigured. Fix the service to enable DNS resolution.`
}

func (l *HeadlessServiceDNS) Hints() []string {
	return []string{
		"Check the headless service configuration",
		"Verify clusterIP is set to None",
		"Test DNS resolution from a pod",
	}
}

func (l *HeadlessServiceDNS) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *HeadlessServiceDNS) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Service
metadata:
  name: headless-svc
spec:
  clusterIP: 10.0.0.100
  selector:
    app: stateful
  ports:
  - port: 80
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  replicas: 3
  serviceName: headless-svc
  selector:
    matchLabels:
      app: stateful
  template:
    metadata:
      labels:
        app: stateful
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *HeadlessServiceDNS) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "headless-svc",
		"-o", "jsonpath={.spec.clusterIP}")
	if err != nil {
		return err
	}
	if output != "None" {
		return fmt.Errorf("service clusterIP is not None: %s", output)
	}
	return nil
}

func (l *HeadlessServiceDNS) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check service", Command: "kubectl get svc headless-svc -o yaml"},
		{Description: "Fix clusterIP", Command: "kubectl patch svc headless-svc -p '{\"spec\":{\"clusterIP\":\"None\"}}'"},
		{Description: "Test DNS", Command: "kubectl exec web-0 -- nslookup web-0.headless-svc.default.svc.cluster.local"},
	}
}
