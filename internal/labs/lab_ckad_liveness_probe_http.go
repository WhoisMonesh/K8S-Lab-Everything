package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADLivenessProbeHTTPLab{})
}

type CKADLivenessProbeHTTPLab struct {
	BaseLab
}

func (l *CKADLivenessProbeHTTPLab) ID() string             { return "ckad_liveness_probe_http" }
func (l *CKADLivenessProbeHTTPLab) Title() string          { return "Configure HTTP-based Liveness Probe" }
func (l *CKADLivenessProbeHTTPLab) Category() Category     { return CategoryAppObservability }
func (l *CKADLivenessProbeHTTPLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADLivenessProbeHTTPLab) Cert() Cert             { return CertCKAD }
func (l *CKADLivenessProbeHTTPLab) DomainWeight() int      { return 15 }
func (l *CKADLivenessProbeHTTPLab) EstimatedTime() int     { return 15 }
func (l *CKADLivenessProbeHTTPLab) Tags() []string {
	return []string{"liveness-probe", "http", "health-check"}
}

func (l *CKADLivenessProbeHTTPLab) Description() string {
	return `A web application needs an HTTP-based liveness probe to check if it's
responding correctly on the /healthz endpoint.

Your task: Add an HTTP liveness probe to the deployment.`
}

func (l *CKADLivenessProbeHTTPLab) Hints() []string {
	return []string{
		"Use livenessProbe with httpGet action",
		"Set the path to /healthz and port to 80",
		"Configure initialDelaySeconds and periodSeconds",
	}
}

func (l *CKADLivenessProbeHTTPLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADLivenessProbeHTTPLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
spec:
  replicas: 2
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: webapp
        image: nginx:alpine
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *CKADLivenessProbeHTTPLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp",
		"-o", "jsonpath={.spec.template.spec.containers[0].livenessProbe.httpGet.path}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no HTTP liveness probe configured")
	}
	return nil
}

func (l *CKADLivenessProbeHTTPLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit deployment", Command: "kubectl edit deployment webapp"},
		{Description: "Add liveness probe", Command: "Add livenessProbe with httpGet path /healthz port 80"},
		{Description: "Verify probe", Command: "kubectl get deployment webapp -o yaml | grep -A 5 livenessProbe"},
	}
}
