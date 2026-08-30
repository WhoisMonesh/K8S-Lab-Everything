package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&ServiceHeadlessLab{})
}

type ServiceHeadlessLab struct {
	BaseLab
}

func (l *ServiceHeadlessLab) ID() string             { return "cka_service_headless" }
func (l *ServiceHeadlessLab) Title() string          { return "Create Headless Service for StatefulSet" }
func (l *ServiceHeadlessLab) Category() Category     { return CategoryServicesNetworking }
func (l *ServiceHeadlessLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *ServiceHeadlessLab) EstimatedTime() int     { return 20 }
func (l *ServiceHeadlessLab) Tags() []string {
	return []string{"headless", "service", "statefulset", "dns"}
}
func (l *ServiceHeadlessLab) Cert() Cert        { return CertCKA }
func (l *ServiceHeadlessLab) DomainWeight() int { return 20 }

func (l *ServiceHeadlessLab) Description() string {
	return `A StatefulSet needs a headless service for stable network identities.
Create a headless service with clusterIP: None that allows direct DNS
lookup of individual pods.`
}

func (l *ServiceHeadlessLab) Hints() []string {
	return []string{
		"Set clusterIP to None for headless service",
		"Use the same selector as the StatefulSet",
		"Pods will get DNS entries like pod-name.service-name.namespace.svc.cluster.local",
	}
}

func (l *ServiceHeadlessLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceHeadlessLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: headless-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: headless-ns
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
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
  name: nginx-headless
  namespace: headless-ns
spec:
  type: ClusterIP
  selector:
    app: nginx
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, svc); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}

func (l *ServiceHeadlessLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *ServiceHeadlessLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "nginx-headless",
		"-n", "headless-ns", "-o", "jsonpath={.spec.clusterIP}")
	if err != nil {
		return err
	}
	if output != "None" {
		return fmt.Errorf("service is not headless (clusterIP != None)")
	}
	return nil
}

func (l *ServiceHeadlessLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create headless service", Command: "cat <<EOF | kubectl apply -f -\napiVersion: v1\nkind: Service\nmetadata:\n  name: nginx-headless\n  namespace: headless-ns\nspec:\n  clusterIP: None\n  selector:\n    app: nginx\n  ports:\n  - port: 80\n    targetPort: 80\nEOF"},
		{Description: "Verify headless", Command: "kubectl get svc nginx-headless -n headless-ns"},
		{Description: "Test DNS lookup", Command: "kubectl exec -n headless-ns <pod> -- nslookup nginx-headless.headless-ns.svc.cluster.local"},
	}
}
