package labs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func init() {
	Register(&PSAEnforcementLab{})
}

type PSAEnforcementLab struct {
	BaseLab
}

func (l *PSAEnforcementLab) ID() string {
	return "psa_enforcement"
}

func (l *PSAEnforcementLab) Title() string {
	return "Pod Security Admission Enforcement"
}

func (l *PSAEnforcementLab) Category() Category {
	return CategoryMicroserviceVulns
}

func (l *PSAEnforcementLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PSAEnforcementLab) Description() string {
	return `A namespace has Pod Security Admission (PSA) enforcing the restricted
policy. A deployment is failing because pods violate the admission
control policy.

Your task: Modify the deployment to comply with the restricted PSA
policy so pods can be admitted and scheduled.`
}

func (l *PSAEnforcementLab) Hints() []string {
	return []string{
		"Check the namespace PSA labels",
		"Restricted policy requires specific security contexts",
		"Pods must run as non-root with dropped capabilities",
	}
}

func (l *PSAEnforcementLab) EstimatedTime() int {
	return 20
}

func (l *PSAEnforcementLab) Tags() []string {
	return []string{"psa", "pod-security", "admission", "security", "namespace"}
}

func (l *PSAEnforcementLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: psa-secure
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
`
	return kubectlApply(ctx, kubeconfigPath, namespace)
}

func (l *PSAEnforcementLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: psa-app
  namespace: psa-secure
spec:
  replicas: 2
  selector:
    matchLabels:
      app: psa-app
  template:
    metadata:
      labels:
        app: psa-app
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

func (l *PSAEnforcementLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "events", "-n", "psa-secure",
		"--field-selector", "reason=FailedCreate",
		"-o", "jsonpath={.items[*].message}")
	if err != nil {
		return nil
	}

	if strings.Contains(output, "forbidden") || strings.Contains(output, "violates") {
		return nil
	}

	pods, _ := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "psa-secure",
		"-o", "jsonpath={.items[*].status.phase}")
	if !strings.Contains(pods, "Running") {
		return nil
	}

	return fmt.Errorf("pods are running (expected PSA rejection)")
}

func (l *PSAEnforcementLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "psa-app",
		"-n", "psa-secure", "-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("checking deployment: %w", err)
	}

	ready, _ := strconv.Atoi(strings.TrimSpace(output))
	if ready < 2 {
		return fmt.Errorf("deployment not ready (ready: %d, expected: 2)", ready)
	}

	return nil
}

func (l *PSAEnforcementLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check namespace PSA labels",
			Command:     "kubectl get namespace psa-secure --show-labels",
			Notes:       "Enforce policy is 'restricted'",
		},
		{
			Description: "Check deployment status",
			Command:     "kubectl get deploy psa-app -n psa-secure",
			Notes:       "Deployment may show 0 ready replicas",
		},
		{
			Description: "Fix: Update security context for restricted PSA",
			Command:     `kubectl patch deploy psa-app -n psa-secure --type='json' -p='[{"op":"replace","path":"/spec/template/spec/containers/0/securityContext","value":{"allowPrivilegeEscalation":false,"runAsNonRoot":true,"capabilities":{"drop":["ALL"]}}}]'`,
			Notes:       "Comply with restricted PSA: no privilege escalation, run as non-root, drop all caps",
		},
		{
			Description: "Verify pods are admitted and running",
			Command:     "kubectl rollout status deploy/psa-app -n psa-secure",
			Notes:       "Pods should now be created successfully",
		},
	}
}
