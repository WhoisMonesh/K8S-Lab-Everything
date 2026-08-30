package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodSecurityStandardsLab{})
}

type PodSecurityStandardsLab struct {
	BaseLab
}

func (l *PodSecurityStandardsLab) ID() string {
	return "pod_security_standards"
}

func (l *PodSecurityStandardsLab) Title() string {
	return "Pod Security Standards Violation"
}

func (l *PodSecurityStandardsLab) Category() Category {
	return CategoryAppConfigSecurity
}

func (l *PodSecurityStandardsLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodSecurityStandardsLab) Description() string {
	return `A namespace has Pod Security Standards enforced at the restricted level.
A deployment is failing because its pods violate the security policy.

Your task: Modify the deployment to comply with the restricted Pod Security Standard.`
}

func (l *PodSecurityStandardsLab) Hints() []string {
	return []string{
		"Check the namespace labels for Pod Security Standards",
		"Restricted standard requires specific security contexts",
		"Pods must run as non-root and drop all capabilities",
	}
}

func (l *PodSecurityStandardsLab) EstimatedTime() int {
	return 20
}

func (l *PodSecurityStandardsLab) Tags() []string {
	return []string{"pod-security", "security-standards", "namespace", "security"}
}

func (l *PodSecurityStandardsLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: secure-ns
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
`
	return kubectlApply(ctx, kubeconfigPath, namespace)
}

func (l *PodSecurityStandardsLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: insecure-app
  namespace: secure-ns
spec:
  replicas: 2
  selector:
    matchLabels:
      app: insecure-app
  template:
    metadata:
      labels:
        app: insecure-app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo running; sleep 15; done']
        securityContext:
          allowPrivilegeEscalation: true
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *PodSecurityStandardsLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "secure-ns",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("checking pods: %w", err)
	}

	if strings.Contains(output, "Running") {
		return fmt.Errorf("pods are running (expected blocked by PSS)")
	}

	return nil
}

func (l *PodSecurityStandardsLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "insecure-app",
		"-n", "secure-ns", "-o", "jsonpath={.spec.template.spec.containers[0].securityContext}")
	if err != nil {
		return fmt.Errorf("checking deployment: %w", err)
	}

	if strings.Contains(output, "allowPrivilegeEscalation: true") {
		return fmt.Errorf("deployment still has allowPrivilegeEscalation: true")
	}

	return nil
}

func (l *PodSecurityStandardsLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check namespace PSS labels",
			Command:     "kubectl get namespace secure-ns --show-labels",
			Notes:       "Enforce policy is 'restricted'",
		},
		{
			Description: "Check deployment security context",
			Command:     "kubectl get deploy insecure-app -n secure-ns -o yaml | grep -A 5 securityContext",
			Notes:       "allowPrivilegeEscalation is true - violates restricted standard",
		},
		{
			Description: "Fix: Update security context for restricted standard",
			Command:     `kubectl patch deploy insecure-app -n secure-ns --type='json' -p='[{"op":"replace","path":"/spec/template/spec/containers/0/securityContext","value":{"allowPrivilegeEscalation":false,"runAsNonRoot":true,"capabilities":{"drop":["ALL"]}}}]'`,
			Notes:       "Comply with restricted PSS: no privilege escalation, run as non-root, drop all caps",
		},
		{
			Description: "Verify pods are running",
			Command:     "kubectl rollout status deploy/insecure-app -n secure-ns",
			Notes:       "Pods should now be created and running",
		},
	}
}
