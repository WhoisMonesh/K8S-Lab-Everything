package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&PodCrashLoopLab{})
}

type PodCrashLoopLab struct {
	BaseLab
}

func (l *PodCrashLoopLab) ID() string {
	return "pod_crashloop"
}

func (l *PodCrashLoopLab) Title() string {
	return "Pod in CrashLoopBackOff"
}

func (l *PodCrashLoopLab) Category() Category {
	return CategoryWorkloads
}

func (l *PodCrashLoopLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *PodCrashLoopLab) Description() string {
	return `A deployment named 'webapp' in the default namespace is failing to start.
The pods are stuck in CrashLoopBackOff state.

Your task: Debug and fix the deployment so the pods run successfully.`
}

func (l *PodCrashLoopLab) Hints() []string {
	return []string{
		"Check the pod status and events",
		"Look at the pod logs to see why it's crashing",
		"Check the container image and command configuration",
		"Environment variables might be incorrectly configured",
	}
}

func (l *PodCrashLoopLab) EstimatedTime() int {
	return 15
}

func (l *PodCrashLoopLab) Tags() []string {
	return []string{"pods", "crashloop", "configmap", "troubleshooting", "workloads"}
}

func (l *PodCrashLoopLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodCrashLoopLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create a deployment with a bad configuration that causes crash loop
	brokenDeployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
  namespace: default
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
      - name: web
        image: nginx:alpine
        command: ["/bin/sh"]
        args: ["-c", "echo Starting && cat /config/app.conf && nginx -g 'daemon off;'"]
        volumeMounts:
        - name: config
          mountPath: /config
      volumes:
      - name: config
        configMap:
          name: webapp-config-that-does-not-exist
`

	if err := kubectlApply(ctx, kubeconfigPath, brokenDeployment); err != nil {
		return fmt.Errorf("creating broken deployment: %w", err)
	}

	return nil
}

func (l *PodCrashLoopLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Wait for pods to enter crash loop
	time.Sleep(15 * time.Second)
	return nil
}

func (l *PodCrashLoopLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if the deployment has all replicas ready
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	if output != "2" {
		return fmt.Errorf("deployment not fully ready yet (ready replicas: %s, expected: 2)", output)
	}

	// Also verify pods are actually running
	output, err = kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=webapp",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pods: %w", err)
	}

	// All pods should be Running
	if output != "Running Running" {
		return fmt.Errorf("not all pods are running yet")
	}

	return nil
}

func (l *PodCrashLoopLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check the deployment status",
			Command:     "kubectl get deployment webapp",
			Notes:       "Notice that not all replicas are ready",
		},
		{
			Description: "Check the pod status",
			Command:     "kubectl get pods -l app=webapp",
			Notes:       "Pods should be in CrashLoopBackOff or Error state",
		},
		{
			Description: "Describe a pod to see events",
			Command:     "kubectl describe pod -l app=webapp | grep -A 10 Events",
			Notes:       "Look for warning messages about missing ConfigMap or mount failures",
		},
		{
			Description: "Check pod logs",
			Command:     "kubectl logs -l app=webapp",
			Notes:       "The logs will show the container failing to start due to missing /config/app.conf",
		},
		{
			Description: "Identify the issue",
			Notes:       "The deployment references a ConfigMap 'webapp-config-that-does-not-exist' which doesn't exist",
		},
		{
			Description: "Create the missing ConfigMap",
			Command:     "kubectl create configmap webapp-config --from-literal=app.conf='server { listen 80; }'",
			Notes:       "This creates the ConfigMap that the deployment expects",
		},
		{
			Description: "Update the deployment to use the correct ConfigMap name",
			Command:     "kubectl patch deployment webapp --type='json' -p='[{\"op\": \"replace\", \"path\": \"/spec/template/spec/volumes/0/configMap/name\", \"value\":\"webapp-config\"}]'",
			Notes:       "This updates the deployment to reference the newly created ConfigMap",
		},
		{
			Description: "Verify the pods are now running",
			Command:     "kubectl get pods -l app=webapp",
			Notes:       "All pods should now be in Running state",
		},
		{
			Description: "Alternative fix: Remove the ConfigMap volume entirely",
			Command:     "kubectl edit deployment webapp",
			Notes:       "You could also remove the volume mount and simplify the container command to just run nginx normally",
		},
	}
}
