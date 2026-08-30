package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&RuntimeClassSandboxLab{})
}

type RuntimeClassSandboxLab struct {
	BaseLab
}

func (l *RuntimeClassSandboxLab) ID() string {
	return "runtimeclass_sandbox"
}

func (l *RuntimeClassSandboxLab) Title() string {
	return "RuntimeClass with Sandbox Container"
}

func (l *RuntimeClassSandboxLab) Category() Category {
	return CategoryMicroserviceVulns
}

func (l *RuntimeClassSandboxLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *RuntimeClassSandboxLab) Description() string {
	return `A RuntimeClass has been created for a sandboxed container runtime,
but the deployment is not using it. Pods should run with the
sandbox RuntimeClass for isolation.

Your task: Update the deployment to use the sandbox RuntimeClass.`
}

func (l *RuntimeClassSandboxLab) Hints() []string {
	return []string{
		"Check available RuntimeClasses",
		"RuntimeClass is specified in pod spec under runtimeClassName",
		"Update the deployment template to include runtimeClassName",
	}
}

func (l *RuntimeClassSandboxLab) EstimatedTime() int {
	return 15
}

func (l *RuntimeClassSandboxLab) Tags() []string {
	return []string{"runtimeclass", "sandbox", "security", "isolation", "container-runtime"}
}

func (l *RuntimeClassSandboxLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	rc := `apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: sandbox
handler: runc
overhead:
  podFixed:
    memory: "128Mi"
    cpu: "250m"
scheduling:
  nodeSelector:
    node-role.kubernetes.io/worker: ""
`
	return kubectlApply(ctx, kubeconfigPath, rc)
}

func (l *RuntimeClassSandboxLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: sandbox-app
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: sandbox-app
  template:
    metadata:
      labels:
        app: sandbox-app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ['sh', '-c', 'while true; do echo running; sleep 15; done']
        resources:
          limits:
            memory: 32Mi
            cpu: 50m
`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *RuntimeClassSandboxLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "sandbox-app",
		"-o", "jsonpath={.spec.template.spec.runtimeClassName}")
	if err != nil {
		return nil
	}

	if strings.TrimSpace(output) == "" {
		return nil
	}

	return fmt.Errorf("deployment has runtimeClassName (expected none)")
}

func (l *RuntimeClassSandboxLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "sandbox-app",
		"-o", "jsonpath={.spec.template.spec.runtimeClassName}")
	if err != nil {
		return fmt.Errorf("checking deployment: %w", err)
	}

	if strings.TrimSpace(output) != "sandbox" {
		return fmt.Errorf("deployment runtimeClassName not set to 'sandbox'")
	}

	return nil
}

func (l *RuntimeClassSandboxLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check available RuntimeClasses",
			Command:     "kubectl get runtimeclass",
			Notes:       "Should see 'sandbox' RuntimeClass",
		},
		{
			Description: "Fix: Add runtimeClassName to deployment",
			Command:     `kubectl patch deploy sandbox-app --type='json' -p='[{"op":"add","path":"/spec/template/spec/runtimeClassName","value":"sandbox"}]'`,
			Notes:       "Set RuntimeClass to sandbox",
		},
		{
			Description: "Verify deployment uses sandbox RuntimeClass",
			Command:     "kubectl get deploy sandbox-app -o jsonpath='{.spec.template.spec.runtimeClassName}'",
			Notes:       "Should return 'sandbox'",
		},
	}
}
