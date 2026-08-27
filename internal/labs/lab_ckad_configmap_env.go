package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADConfigMapEnvLab{})
}

type CKADConfigMapEnvLab struct {
	BaseLab
}

func (l *CKADConfigMapEnvLab) ID() string             { return "ckad_configmap_env" }
func (l *CKADConfigMapEnvLab) Title() string          { return "Use ConfigMap for Environment Variables" }
func (l *CKADConfigMapEnvLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADConfigMapEnvLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADConfigMapEnvLab) Cert() Cert             { return CertCKAD }
func (l *CKADConfigMapEnvLab) DomainWeight() int      { return 25 }
func (l *CKADConfigMapEnvLab) EstimatedTime() int     { return 15 }
func (l *CKADConfigMapEnvLab) Tags() []string {
	return []string{"configmap", "env", "environment"}
}

func (l *CKADConfigMapEnvLab) Description() string {
	return `A pod needs environment variables sourced from a ConfigMap. The ConfigMap
'env-config' exists but the pod isn't using it.

Your task: Update the pod to use the ConfigMap for environment variables.`
}

func (l *CKADConfigMapEnvLab) Hints() []string {
	return []string{
		"Use envFrom with configMapRef to load all keys",
		"Use env with valueFrom and configMapKeyRef for specific keys",
		"Each ConfigMap key becomes an environment variable",
	}
}

func (l *CKADConfigMapEnvLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADConfigMapEnvLab) Break(ctx context.Context, kubeconfigPath string) error {
	cm := `apiVersion: v1
kind: ConfigMap
metadata:
  name: env-config
data:
  DATABASE_HOST: mysql.default.svc
  DATABASE_PORT: "3306"
  APP_ENV: production`
	if err := kubectlApply(ctx, kubeconfigPath, cm); err != nil {
		return fmt.Errorf("creating configmap: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: env-app
  labels:
    app: env-app
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'env | sort && sleep 3600']`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADConfigMapEnvLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "env-app",
		"-o", "jsonpath={.spec.containers[0].envFrom[*].configMapRef.name}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no envFrom configMapRef found")
	}
	if !strings.Contains(output, "env-config") {
		return fmt.Errorf("ConfigMap 'env-config' not referenced")
	}
	return nil
}

func (l *CKADConfigMapEnvLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check ConfigMap", Command: "kubectl get configmap env-config -o yaml"},
		{Description: "Edit pod", Command: "kubectl edit pod env-app"},
		{Description: "Add envFrom", Command: "Add envFrom with configMapRef name: env-config"},
		{Description: "Verify env vars", Command: "kubectl exec env-app -- env | grep DATABASE"},
	}
}
