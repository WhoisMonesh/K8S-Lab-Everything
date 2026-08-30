package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&NginxIngressTLSLab{})
}

type NginxIngressTLSLab struct {
	BaseLab
}

func (l *NginxIngressTLSLab) ID() string {
	return "nginx_ingress_tls"
}

func (l *NginxIngressTLSLab) Title() string {
	return "Nginx Deployment with Ingress and TLS"
}

func (l *NginxIngressTLSLab) Category() Category {
	return CategoryServicesNetworkCKAD
}

func (l *NginxIngressTLSLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *NginxIngressTLSLab) Description() string {
	return `Deploy an nginx application with a Service and Ingress resource.
The Ingress is misconfigured with incorrect backend service references.

Your task: Fix the Ingress configuration so traffic routes correctly
to the nginx backend service.`
}

func (l *NginxIngressTLSLab) Hints() []string {
	return []string{
		"Check the Ingress resource and its backend configuration",
		"Verify the Service name and port match the Ingress backend",
		"Use kubectl describe ingress to see routing rules",
	}
}

func (l *NginxIngressTLSLab) EstimatedTime() int {
	return 20
}

func (l *NginxIngressTLSLab) Tags() []string {
	return []string{"nginx", "ingress", "tls", "deployment", "service", "real-app"}
}

func (l *NginxIngressTLSLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NginxIngressTLSLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-app
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx-app
  template:
    metadata:
      labels:
        app: nginx-app
    spec:
      containers:
      - name: nginx
        image: nginx:1.25
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return err
	}

	service := `apiVersion: v1
kind: Service
metadata:
  name: nginx-svc
  namespace: default
spec:
  selector:
    app: nginx-app
  ports:
  - port: 80
    targetPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, service); err != nil {
		return err
	}

	ingress := `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx-ingress
  namespace: default
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: nginx.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nginx-wrong-svc
            port:
              number: 8080
`
	return kubectlApply(ctx, kubeconfigPath, ingress)
}

func (l *NginxIngressTLSLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "nginx-ingress",
		"-o", "jsonpath={.spec.rules[0].http.paths[0].backend.service.name}")
	if err != nil {
		return nil
	}

	if strings.TrimSpace(output) == "nginx-wrong-svc" {
		return nil
	}

	return fmt.Errorf("ingress backend is correct (expected wrong)")
}

func (l *NginxIngressTLSLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "nginx-ingress",
		"-o", "jsonpath={.spec.rules[0].http.paths[0].backend.service.name}")
	if err != nil {
		return fmt.Errorf("checking ingress: %w", err)
	}

	if strings.TrimSpace(output) == "nginx-wrong-svc" {
		return fmt.Errorf("ingress still points to wrong service")
	}

	return nil
}

func (l *NginxIngressTLSLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check Ingress configuration",
			Command:     "kubectl get ingress nginx-ingress -o yaml",
			Notes:       "Backend points to nginx-wrong-svc:8080",
		},
		{
			Description: "Fix: Update Ingress backend",
			Command:     `kubectl patch ingress nginx-ingress --type='json' -p='[{"op":"replace","path":"/spec/rules/0/http/paths/0/backend/service/name","value":"nginx-svc"},{"op":"replace","path":"/spec/rules/0/http/paths/0/backend/service/port/number","value":80}]'`,
			Notes:       "Point to correct service name and port",
		},
		{
			Description: "Verify Ingress routing",
			Command:     "kubectl describe ingress nginx-ingress",
			Notes:       "Backend should now point to nginx-svc:80",
		},
	}
}
