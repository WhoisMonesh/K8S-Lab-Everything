package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADEphemeralContainerDebugLab{})
}

type CKADEphemeralContainerDebugLab struct {
	BaseLab
}

func (l *CKADEphemeralContainerDebugLab) ID() string             { return "ckad_ephemeral_container_debug" }
func (l *CKADEphemeralContainerDebugLab) Title() string          { return "Debug Using Ephemeral Containers" }
func (l *CKADEphemeralContainerDebugLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADEphemeralContainerDebugLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADEphemeralContainerDebugLab) Cert() Cert             { return CertCKAD }
func (l *CKADEphemeralContainerDebugLab) DomainWeight() int      { return 20 }
func (l *CKADEphemeralContainerDebugLab) EstimatedTime() int     { return 20 }
func (l *CKADEphemeralContainerDebugLab) Tags() []string {
	return []string{"ephemeral", "debugging", "troubleshooting"}
}

func (l *CKADEphemeralContainerDebugLab) Description() string {
	return `A production pod is running but the application is misbehaving.
You cannot restart the pod without disrupting service.

Your task: Add an ephemeral debug container to investigate the issue
without restarting the pod.`
}

func (l *CKADEphemeralContainerDebugLab) Hints() []string {
	return []string{
		"Use kubectl debug to add an ephemeral container",
		"Check the ephemeral container status",
		"Execute commands inside the debug container",
	}
}

func (l *CKADEphemeralContainerDebugLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADEphemeralContainerDebugLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: debug-target
spec:
  replicas: 1
  selector:
    matchLabels:
      app: debug-target
  template:
    metadata:
      labels:
        app: debug-target
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *CKADEphemeralContainerDebugLab) Verify(ctx context.Context, kubeconfigPath string) error {
	pods, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=debug-target",
		"-o", "jsonpath={.items[0].metadata.name}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", strings.TrimSpace(pods),
		"-o", "jsonpath={.status.ephemeralContainerStatuses[*].state}")
	if err != nil {
		return fmt.Errorf("failed to check ephemeral containers: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no ephemeral container found")
	}
	return nil
}

func (l *CKADEphemeralContainerDebugLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Get pod name", Command: "kubectl get pods -l app=debug-target"},
		{Description: "Add ephemeral container", Command: "kubectl debug -it <pod-name> --image=busybox:1.36 --target=app"},
		{Description: "Debug the application", Command: "Use nslookup, wget, or other tools to investigate"},
	}
}
