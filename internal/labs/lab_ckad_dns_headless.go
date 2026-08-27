package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADDNSHeadlessLab{})
}

type CKADDNSHeadlessLab struct {
	BaseLab
}

func (l *CKADDNSHeadlessLab) ID() string             { return "ckad_dns_headless" }
func (l *CKADDNSHeadlessLab) Title() string          { return "Configure Headless Service DNS" }
func (l *CKADDNSHeadlessLab) Category() Category     { return CategoryServicesNetworkCKAD }
func (l *CKADDNSHeadlessLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADDNSHeadlessLab) Cert() Cert             { return CertCKAD }
func (l *CKADDNSHeadlessLab) DomainWeight() int      { return 20 }
func (l *CKADDNSHeadlessLab) EstimatedTime() int     { return 15 }
func (l *CKADDNSHeadlessLab) Tags() []string {
	return []string{"dns", "headless", "statefulset"}
}

func (l *CKADDNSHeadlessLab) Description() string {
	return `A StatefulSet needs individual pod DNS entries. Create a headless service
that provides unique DNS records for each pod.

Your task: Create a headless service for the StatefulSet.`
}

func (l *CKADDNSHeadlessLab) Hints() []string {
	return []string{
		"Set clusterIP: None for a headless service",
		"Each pod gets a DNS entry like pod-name.service-name.namespace.svc.cluster.local",
		"Headless services don't load balance",
	}
}

func (l *CKADDNSHeadlessLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADDNSHeadlessLab) Break(ctx context.Context, kubeconfigPath string) error {
	statefulset := `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web
spec:
  replicas: 2
  serviceName: web-headless
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, statefulset)
}

func (l *CKADDNSHeadlessLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "web-headless",
		"-o", "jsonpath={.spec.clusterIP}")
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}
	if strings.TrimSpace(output) != "None" {
		return fmt.Errorf("service is not headless (clusterIP: %s)", output)
	}
	return nil
}

func (l *CKADDNSHeadlessLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create headless service", Command: "Create service with clusterIP: None"},
		{Description: "Verify", Command: "kubectl get service web-headless -o yaml | grep clusterIP"},
		{Description: "Check DNS", Command: "kubectl exec web-0 -- nslookup web-0.web-headless.default.svc.cluster.local"},
	}
}
