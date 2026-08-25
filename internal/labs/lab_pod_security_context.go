package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodSecurityContextLab{})
}

type PodSecurityContextLab struct {
	BaseLab
}

func (l *PodSecurityContextLab) ID() string {
	return "pod_security_context"
}

func (l *PodSecurityContextLab) Title() string {
	return "Pod Failing Due to Security Context"
}

func (l *PodSecurityContextLab) Category() Category {
	return CategorySecurity
}

func (l *PodSecurityContextLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodSecurityContextLab) Description() string {
	return `A pod 'restricted-app' is failing to start because it's trying to run as root but the PodSecurityPolicy/SecurityContext forbids it.

Your task: Fix the security context to allow the pod to run correctly.`
}

func (l *PodSecurityContextLab) Hints() []string {
	return []string{
		"Check the pod status and events",
		"Look at the securityContext in the pod spec",
		"Check runAsNonRoot and runAsUser settings",
		"The container image might need to run as a specific user",
	}
}

func (l *PodSecurityContextLab) EstimatedTime() int {
	return 15
}

func (l *PodSecurityContextLab) Tags() []string {
	return []string{"security", "security-context", "runasnonroot", "pod-security"}
}

func (l *PodSecurityContextLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodSecurityContextLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create pod with restrictive security context
	// nginx:alpine runs as root by default, but we're forcing runAsNonRoot=true
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: restricted-app
  namespace: default
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 0
  containers:
  - name: web
    image: nginx:alpine
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
    volumeMounts:
    - name: cache
      mountPath: /var/cache/nginx
    - name: run
      mountPath: /var/run
  volumes:
  - name: cache
    emptyDir: {}
  - name: run
    emptyDir: {}
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodSecurityContextLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodSecurityContextLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if pod is running
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "restricted-app",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *PodSecurityContextLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod restricted-app",
			Notes:       "The pod should be in CreateError or CrashLoopBackOff",
		},
		{
			Description: "Check pod events",
			Command:     "kubectl describe pod restricted-app | grep -A 10 Events",
			Notes:       "Look for 'runAsNonRoot set to true but container runs as root' error",
		},
		{
			Description: "Identify the issue",
			Notes:       "runAsUser: 0 is root, but runAsNonRoot: true prevents running as root",
		},
		{
			Description: "Fix the security context",
			Command:     "kubectl delete pod restricted-app",
			Notes:       "Delete the pod to recreate with corrected security context",
		},
		{
			Description: "Create pod with correct security context",
			Command:     `kubectl run restricted-app --image=nginx:alpine --dry-run=client -o yaml > fixed-pod.yaml`,
			Notes:       "Edit the YAML to set runAsUser: 101 (nginx user) and runAsNonRoot: true",
		},
		{
			Description: "Apply the fixed pod",
			Command:     "kubectl apply -f fixed-pod.yaml",
			Notes:       "The pod should now start successfully",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod restricted-app",
			Notes:       "The pod should now be in Running state",
		},
	}
}
