package labs

import (
    "context"
    "fmt"
    "strconv"
    "strings"
    "time"
)

func init() {
    Register(&PSABaselineLab{})
}

type PSABaselineLab struct {
    BaseLab
}

func (l *PSABaselineLab) ID() string { return "psa_baseline" }
func (l *PSABaselineLab) Title() string { return "Pod Security Admission: Baseline" }
func (l *PSABaselineLab) Category() Category { return CategoryMicroserviceVulns }
func (l *PSABaselineLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PSABaselineLab) Description() string {
    return `A namespace has Pod Security Admission (PSA) enforcing the baseline policy. A deployment fails because pods run as root, which is not allowed under baseline.

Your task: Modify the deployment to comply with the baseline PSA policy so pods can be admitted and scheduled.`
}
func (l *PSABaselineLab) Hints() []string {
    return []string{"Check the namespace PSA labels", "Baseline policy disallows running as root", "Set runAsNonRoot: true in the pod securityContext"}
}
func (l *PSABaselineLab) EstimatedTime() int { return 20 }
func (l *PSABaselineLab) Tags() []string { return []string{"psa", "pod-security", "baseline", "security", "namespace"} }

func (l *PSABaselineLab) Prepare(ctx context.Context, kubeconfigPath string) error {
    if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
        return err
    }
    ns := `apiVersion: v1
kind: Namespace
metadata:
  name: psa-baseline
  labels:
    pod-security.kubernetes.io/enforce: baseline
    pod-security.kubernetes.io/audit: baseline
    pod-security.kubernetes.io/warn: baseline`
    return kubectlApply(ctx, kubeconfigPath, ns)
}

func (l *PSABaselineLab) Break(ctx context.Context, kubeconfigPath string) error {
    deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: psa-app
  namespace: psa-baseline
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
          runAsUser: 0  # runs as root, violates baseline
`
    return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *PSABaselineLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
    time.Sleep(10 * time.Second)
    output, err := kubectl(ctx, kubeconfigPath, "get", "events", "-n", "psa-baseline",
        "--field-selector", "reason=FailedCreate", "-o", "jsonpath={.items[*].message}")
    if err != nil {
        return nil
    }
    if strings.Contains(output, "running as root") || strings.Contains(output, "allowed to run as root") {
        return nil
    }
    // If pods are running, it's unexpected
    pods, _ := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "psa-baseline",
        "-o", "jsonpath={.items[*].status.phase}")
    if !strings.Contains(pods, "Running") {
        return nil
    }
    return fmt.Errorf("pods are running (expected PSA rejection)")
}

func (l *PSABaselineLab) Verify(ctx context.Context, kubeconfigPath string) error {
    time.Sleep(10 * time.Second)
    output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "psa-app",
        "-n", "psa-baseline", "-o", "jsonpath={.status.readyReplicas}")
    if err != nil {
        return fmt.Errorf("checking deployment: %w", err)
    }
    ready, _ := strconv.Atoi(strings.TrimSpace(output))
    if ready < 2 {
        return fmt.Errorf("deployment not ready (ready: %d, expected: 2)", ready)
    }
    return nil
}

func (l *PSABaselineLab) SolutionSteps() []SolutionStep {
    return []SolutionStep{{
        Description: "Patch the deployment to set runAsNonRoot: true",
        Command: `kubectl patch deploy psa-app -n psa-baseline --type='json' -p='[{"op":"add","path":"/spec/template/spec/containers/0/securityContext/runAsNonRoot","value":true}]'`,
        Notes: "Baseline PSA requires containers not to run as root",
    }, {
        Description: "Verify the deployment is ready",
        Command: "kubectl rollout status deploy/psa-app -n psa-baseline",
        Notes: "Pods should now be created successfully",
    }}
}
