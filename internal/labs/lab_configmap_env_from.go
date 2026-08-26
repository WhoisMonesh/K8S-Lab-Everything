package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ConfigMapEnvFrom{})
}

type ConfigMapEnvFrom struct {
	BaseLab
}

func (l *ConfigMapEnvFrom) ID() string            { return "configmap_env_from" }
func (l *ConfigMapEnvFrom) Title() string         { return "ConfigMap envFrom Reference Broken" }
func (l *ConfigMapEnvFrom) Category() Category    { return CategoryWorkloads }
func (l *ConfigMapEnvFrom) Difficulty() Difficulty { return DifficultyEasy }
func (l *ConfigMapEnvFrom) EstimatedTime() int    { return 10 }
func (l *ConfigMapEnvFrom) Tags() []string        { return []string{"configmap", "environment", "workloads"} }

func (l *ConfigMapEnvFrom) Description() string {
	return `A pod is failing because it references a ConfigMap via envFrom that doesn't exist.
Create the missing ConfigMap or fix the reference.`
}

func (l *ConfigMapEnvFrom) Hints() []string {
	return []string{
		"Check the pod spec for envFrom",
		"Verify the ConfigMap exists",
		"Create the ConfigMap with required keys",
	}
}

func (l *ConfigMapEnvFrom) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ConfigMapEnvFrom) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: env-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: env-app
  template:
    metadata:
      labels:
        app: env-app
    spec:
      containers:
      - name: app
        image: busybox:1.36
        command: ["sh", "-c", "env && sleep 3600"]
        envFrom:
        - configMapRef:
            name: app-env`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ConfigMapEnvFrom) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/env-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return err
	}
	if output != "1" {
		return fmt.Errorf("deployment not ready")
	}
	return nil
}

func (l *ConfigMapEnvFrom) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod events", Command: "kubectl describe deploy env-app"},
		{Description: "Create ConfigMap", Command: "kubectl create configmap app-env --from-literal=APP_ENV=production"},
		{Description: "Verify pods", Command: "kubectl get pods -l app=env-app"},
	}
}
