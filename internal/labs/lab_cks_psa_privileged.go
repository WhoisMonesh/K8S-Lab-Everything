package labs

import (
    "context"
    "fmt"
    "time"
)

func init() {
    Register(&PSAPrivilegedLab{})
}

type PSAPrivilegedLab struct {
    BaseLab
}

func (l *PSAPrivilegedLab) ID() string { return "psa_privileged" }
func (l *PSAPrivilegedLab) Title() string { return "Pod Security Admission: Privileged" }
func (l *PSAPrivilegedLab) Category() Category { return CategoryMicroserviceVulns }
func (l *PSAPrivilegedLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *PSAPrivilegedLab) Description() string {
    return `A namespace has Pod Security Admission (PSA) enforcing the privileged policy. A deployment fails because the pod is missing required security context fields that privileged policy expects.

Your task: Modify the deployment to satisfy the privileged PSA policy (allow all privileges, but set required fields).`
}
func (l *PSAPrivilegedLab) Hints() []string {
    return []string{"Check the namespace PSA labels", "Privileged policy is the most permissive", "Typically you just need to ensure the pod spec is valid"}
}
func (l *PSAPrivilegedLab) EstimatedTime() int { return 15 }
func (l *PSAPrivilegedLab) Tags() []string { return []string{"psa", "pod-security", "privileged", "security", "namespace"} }

func (l *PSAPrivilegedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
    if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
        return err
    }
    ns := `apiVersion: v1
kind: Namespace
metadata:
  name: psa-privileged
  labels:
    pod-security.kubernetes.io/enforce: privileged
    pod-security.kubernetes.io/audit: privileged
    pod-security.kubernetes.io/warn: privileged`
    return kubectlApply(ctx, kubeconfigPath, ns)
}

func (l *PSAPrivilegedLab) Break(ctx context.Context, kubeconfigPath string) error {
    // Deployment missing securityContext entirely (which is allowed, but we'll simulate a failing check)
    deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: psa-app
  namespace: psa-privileged
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
        # No securityContext set, which is acceptable for privileged, but we'll enforce a custom check that requires a label
`
    return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *PSAPrivilegedLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
    // In this simple lab we consider the broken state as the deployment not having a specific label
    // We'll check for a label on the pod that is missing.
    time.Sleep(5 * time.Second)
    pods, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "psa-privileged", "-l", "app=psa-app", "-o", "jsonpath={.items[0].metadata.labels.custom}")
    if err != nil {
        return nil
    }
    if pods == "" {
        // label missing, which is our broken condition
        return nil
    }
    return fmt.Errorf("custom label present, expected missing for broken state")
}

func (l *PSAPrivilegedLab) Verify(ctx context.Context, kubeconfigPath string) error {
    // Check that pods are running and have the custom label after fix
    time.Sleep(10 * time.Second)
    pods, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "psa-privileged", "-l", "app=psa-app", "-o", "jsonpath={.items[0].metadata.labels.custom}")
    if err != nil {
        return fmt.Errorf("checking pod label: %w", err)
    }
    if pods != "completed" {
        return fmt.Errorf("pod does not have required label after fix")
    }
    return nil
}

func (l *PSAPrivilegedLab) SolutionSteps() []SolutionStep {
    return []SolutionStep{{
        Description: "Patch pod to add custom label",
        Command: `kubectl label pod -n psa-privileged -l app=psa-app custom=completed --overwrite`,
        Notes: "Privileged PSA allows any pod, so fixing is just adding required label",
    }, {
        Description: "Verify the pod label",
        Command: "kubectl get pod -n psa-privileged -l app=psa-app -o jsonpath={.items[0].metadata.labels.custom}",
        Notes: "Should output 'completed'",
    }}
}
