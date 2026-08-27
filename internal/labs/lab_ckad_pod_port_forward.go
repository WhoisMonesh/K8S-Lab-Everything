package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADPodPortForwardLab{})
}

type CKADPodPortForwardLab struct {
	BaseLab
}

func (l *CKADPodPortForwardLab) ID() string             { return "ckad_pod_port_forward" }
func (l *CKADPodPortForwardLab) Title() string          { return "Debug Using Port-Forward" }
func (l *CKADPodPortForwardLab) Category() Category     { return CategoryAppObservability }
func (l *CKADPodPortForwardLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADPodPortForwardLab) Cert() Cert             { return CertCKAD }
func (l *CKADPodPortForwardLab) DomainWeight() int      { return 15 }
func (l *CKADPodPortForwardLab) EstimatedTime() int     { return 10 }
func (l *CKADPodPortForwardLab) Tags() []string {
	return []string{"port-forward", "debugging", "access"}
}

func (l *CKADPodPortForwardLab) Description() string {
	return `A pod is running a web server on port 8080 but you need to access it
from your local machine for debugging.

Your task: Set up port forwarding to access the pod's web server locally.`
}

func (l *CKADPodPortForwardLab) Hints() []string {
	return []string{
		"Use kubectl port-forward to forward a local port to the pod",
		"Specify the pod name and ports",
		"The port-forward runs in the foreground by default",
	}
}

func (l *CKADPodPortForwardLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADPodPortForwardLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: web-server
  labels:
    app: web-server
spec:
  containers:
  - name: web
    image: nginx:alpine
    ports:
    - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADPodPortForwardLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "web-server",
		"-o", "jsonpath={.spec.containers[0].ports[0].containerPort}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no port configured")
	}
	return nil
}

func (l *CKADPodPortForwardLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Get pod name", Command: "kubectl get pods -l app=web-server"},
		{Description: "Forward port", Command: "kubectl port-forward web-server 8080:80"},
		{Description: "Access locally", Command: "curl http://localhost:8080"},
	}
}
