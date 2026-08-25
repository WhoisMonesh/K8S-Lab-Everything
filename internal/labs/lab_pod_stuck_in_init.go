package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodStuckInInitLab{})
}

type PodStuckInInitLab struct {
	BaseLab
}

func (l *PodStuckInInitLab) ID() string {
	return "pod_stuck_in_init"
}

func (l *PodStuckInInitLab) Title() string {
	return "Pod Stuck in Init Container"
}

func (l *PodStuckInInitLab) Category() Category {
	return CategoryWorkloads
}

func (l *PodStuckInInitLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodStuckInInitLab) Description() string {
	return `A pod 'web-app' is stuck in Init:Error state. The init container is failing,
which prevents the main container from starting.

Your task: Fix the init container configuration so the pod can start.`
}

func (l *PodStuckInInitLab) Hints() []string {
	return []string{
		"Check the pod status",
		"Look at the init container status and logs",
		"The init container command might be wrong",
		"Check if the init container image exists",
	}
}

func (l *PodStuckInInitLab) EstimatedTime() int {
	return 15
}

func (l *PodStuckInInitLab) Tags() []string {
	return []string{"init-container", "pod", "troubleshooting", "workloads"}
}

func (l *PodStuckInInitLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodStuckInInitLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create pod with failing init container
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: web-app
  namespace: default
spec:
  initContainers:
  - name: init-check
    image: busybox:99.99-doesnotexist
    command: ['sh', '-c', 'echo "Checking dependencies..." && sleep 5']
  containers:
  - name: web
    image: nginx:alpine
    ports:
    - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodStuckInInitLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *PodStuckInInitLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if pod is running (not in init)
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "web-app",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	// Check init container status
	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "web-app",
		"-o", "jsonpath={.status.initContainerStatuses[0].ready}")
	if err != nil {
		return fmt.Errorf("failed to check init container: %w", err)
	}

	if strings.TrimSpace(output) != "true" {
		return fmt.Errorf("init container is not ready")
	}

	return nil
}

func (l *PodStuckInInitLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod web-app",
			Notes:       "The pod should show Init:Error or Init:CrashLoopBackOff",
		},
		{
			Description: "Check init container status",
			Command:     "kubectl get pod web-app -o jsonpath='{.status.initContainerStatuses}'",
			Notes:       "Look at the state and exit code",
		},
		{
			Description: "Check init container logs",
			Command:     "kubectl logs web-app -c init-check --previous",
			Notes:       "Look for image pull errors or command failures",
		},
		{
			Description: "Identify the issue",
			Notes:       "The init container uses busybox:99.99-doesnotexist which doesn't exist",
		},
		{
			Description: "Delete and recreate pod with correct init container",
			Command:     "kubectl delete pod web-app",
			Notes:       "Delete the broken pod",
		},
		{
			Description: "Create pod with valid init container image",
			Command:     `kubectl run web-app --image=nginx:alpine --dry-run=client -o yaml > fixed-pod.yaml`,
			Notes:       "Edit the YAML to add an init container with busybox:1.28",
		},
		{
			Description: "Apply the fixed pod",
			Command:     "kubectl apply -f fixed-pod.yaml",
			Notes:       "The pod should now start successfully",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod web-app",
			Notes:       "The pod should now be in Running state",
		},
	}
}
