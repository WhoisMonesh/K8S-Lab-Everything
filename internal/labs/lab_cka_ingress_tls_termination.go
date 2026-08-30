package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&IngressTLSTerminationLab{})
}

type IngressTLSTerminationLab struct {
	BaseLab
}

func (l *IngressTLSTerminationLab) ID() string { return "cka_ingress_tls_termination" }
func (l *IngressTLSTerminationLab) Title() string {
	return "Configure TLS Termination on Ingress"
}
func (l *IngressTLSTerminationLab) Category() Category     { return CategoryServicesNetworking }
func (l *IngressTLSTerminationLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *IngressTLSTerminationLab) EstimatedTime() int     { return 20 }
func (l *IngressTLSTerminationLab) Tags() []string {
	return []string{"ingress", "tls", "ssl", "certificates"}
}
func (l *IngressTLSTerminationLab) Cert() Cert        { return CertCKA }
func (l *IngressTLSTerminationLab) DomainWeight() int { return 20 }

func (l *IngressTLSTerminationLab) Description() string {
	return `An Ingress resource is serving traffic without TLS. Configure TLS
termination using a Secret containing TLS certificates and update the
Ingress to use HTTPS.`
}

func (l *IngressTLSTerminationLab) Hints() []string {
	return []string{
		"Create a TLS Secret with tls.crt and tls.key",
		"Add tls section to the Ingress spec",
		"Reference the Secret in the tls section",
	}
}

func (l *IngressTLSTerminationLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *IngressTLSTerminationLab) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: ingress-ns
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: secure-app
  namespace: ingress-ns
spec:
  replicas: 2
  selector:
    matchLabels:
      app: secure-app
  template:
    metadata:
      labels:
        app: secure-app
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: secure-app
  namespace: ingress-ns
spec:
  selector:
    app: secure-app
  ports:
  - port: 80
    targetPort: 80
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: secure-app
  namespace: ingress-ns
spec:
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: secure-app
            port:
              number: 80
`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *IngressTLSTerminationLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *IngressTLSTerminationLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "ingress", "secure-app",
		"-n", "ingress-ns", "-o", "jsonpath={.spec.tls}")
	if err != nil {
		return err
	}
	if output == "" || output == "null" {
		return fmt.Errorf("TLS not configured on Ingress")
	}
	return nil
}

func (l *IngressTLSTerminationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create TLS Secret", Command: "kubectl create secret tls app-tls --cert=tls.crt --key=tls.key -n ingress-ns"},
		{Description: "Add TLS to Ingress", Command: "kubectl patch ingress secure-app -n ingress-ns -p '{\"spec\":{\"tls\":[{\"hosts\":[\"app.example.com\"],\"secretName\":\"app-tls\"}]}}'"},
		{Description: "Verify", Command: "kubectl get ingress secure-app -n ingress-ns -o yaml"},
	}
}
