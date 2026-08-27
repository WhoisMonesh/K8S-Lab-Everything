package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&SchedulerNotRunningLab{})
}

type SchedulerNotRunningLab struct {
	BaseLab
}

func (l *SchedulerNotRunningLab) ID() string {
	return "scheduler_not_running"
}

func (l *SchedulerNotRunningLab) Title() string {
	return "Kube-Scheduler Not Running"
}

func (l *SchedulerNotRunningLab) Category() Category {
	return CategoryScheduling
}

func (l *SchedulerNotRunningLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *SchedulerNotRunningLab) Description() string {
	return `The kube-scheduler is not running properly.
New pods are stuck in the Pending state and are not being scheduled to nodes.

Your task: Fix the kube-scheduler so it can schedule pods again.`
}

func (l *SchedulerNotRunningLab) Hints() []string {
	return []string{
		"Check the kube-scheduler pod status in the kube-system namespace",
		"Look at the kube-scheduler static pod manifest in /etc/kubernetes/manifests",
		"Check for typos or invalid configuration in the manifest",
		"The kubelet automatically restarts static pods when their manifests change",
	}
}

func (l *SchedulerNotRunningLab) EstimatedTime() int {
	return 20
}

func (l *SchedulerNotRunningLab) Tags() []string {
	return []string{"scheduler", "static-pods", "scheduling", "troubleshooting"}
}

func (l *SchedulerNotRunningLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SchedulerNotRunningLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Get the control plane node name
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	containerName := nodeName

	// Read the current kube-scheduler manifest
	output, err := dockerExec(ctx, containerName, "cat", "/etc/kubernetes/manifests/kube-scheduler.yaml")
	if err != nil {
		return fmt.Errorf("reading kube-scheduler manifest: %w", err)
	}

	// Break the scheduler by adding an invalid flag
	// Add a line with invalid syntax after the command section
	modifiedManifest := strings.Replace(output,
		"- kube-scheduler",
		"- kube-scheduler\n    - --invalid-flag-that-does-not-exist=true",
		1)

	// Write the modified manifest back
	writeCmd := fmt.Sprintf("cat > /etc/kubernetes/manifests/kube-scheduler.yaml << 'EOF'\n%s\nEOF", modifiedManifest)
	_, err = dockerExec(ctx, containerName, "sh", "-c", writeCmd)
	if err != nil {
		return fmt.Errorf("writing modified manifest: %w", err)
	}

	return nil
}

func (l *SchedulerNotRunningLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Wait for the scheduler to restart and fail
	time.Sleep(10 * time.Second)

	// Create a test pod to verify scheduling is broken
	testPod := `apiVersion: v1
kind: Pod
metadata:
  name: test-scheduling
  namespace: default
spec:
  containers:
  - name: nginx
    image: nginx:alpine
`
	if err := kubectlApply(ctx, kubeconfigPath, testPod); err != nil {
		// Might fail if API server is having issues, that's okay
		return nil
	}

	// Check if the pod is stuck in Pending
	time.Sleep(5 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pod", "test-scheduling", "-o", "jsonpath={.status.phase}")
	if strings.TrimSpace(output) == "Pending" {
		return nil
	}

	return nil
}

func (l *SchedulerNotRunningLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if the kube-scheduler pod is running
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-scheduler",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check scheduler pod: %w", err)
	}

	if !strings.Contains(output, "Running") {
		return fmt.Errorf("scheduler pod is not running yet")
	}

	// Create a test pod to verify scheduling works
	testPod := `apiVersion: v1
kind: Pod
metadata:
  name: verify-scheduling
  namespace: default
spec:
  containers:
  - name: nginx
    image: nginx:alpine
`
	if err := kubectlApply(ctx, kubeconfigPath, testPod); err != nil {
		return fmt.Errorf("failed to create test pod: %w", err)
	}

	// Wait and check if the pod gets scheduled
	time.Sleep(10 * time.Second)
	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "verify-scheduling",
		"-o", "jsonpath={.spec.nodeName}")
	if err != nil {
		// Clean up the test pod
		kubectl(ctx, kubeconfigPath, "delete", "pod", "verify-scheduling", "--ignore-not-found=true")
		return fmt.Errorf("failed to check test pod: %w", err)
	}

	// Clean up the test pod
	kubectl(ctx, kubeconfigPath, "delete", "pod", "verify-scheduling", "--ignore-not-found=true")

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("test pod was not scheduled to a node")
	}

	return nil
}

func (l *SchedulerNotRunningLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check the scheduler pod status",
			Command:     "kubectl get pods -n kube-system | grep scheduler",
			Notes:       "The kube-scheduler pod should be in CrashLoopBackOff or Error state",
		},
		{
			Description: "Check the scheduler pod logs",
			Command:     "kubectl logs -n kube-system kube-scheduler-<node-name>",
			Notes:       "Look for error messages indicating what's wrong",
		},
		{
			Description: "Access the control plane node",
			Command:     "docker exec -it <cluster-name>-control-plane bash",
		},
		{
			Description: "Examine the kube-scheduler manifest",
			Command:     "cat /etc/kubernetes/manifests/kube-scheduler.yaml",
			Notes:       "Look for invalid flags or configuration errors",
		},
		{
			Description: "Remove the invalid flag",
			Command:     "sed -i '/--invalid-flag-that-does-not-exist/d' /etc/kubernetes/manifests/kube-scheduler.yaml",
			Notes:       "The kubelet will detect the change and restart the scheduler",
		},
		{
			Description: "Verify the scheduler is running",
			Command:     "kubectl get pods -n kube-system | grep scheduler",
			Notes:       "The scheduler pod should now be in Running state",
		},
		{
			Description: "Test pod scheduling",
			Command:     "kubectl run test-nginx --image=nginx:alpine",
			Notes:       "The pod should be scheduled and running",
		},
	}
}
