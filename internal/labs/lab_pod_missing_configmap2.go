package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodMissingConfigMap2{})
}

type PodMissingConfigMap2 struct {
	BaseLab
}

func (l *PodMissingConfigMap2) ID() string            { return "pod_missing_configmap2" }
func (l *PodMissingConfigMap2) Title() string         { return "Pod Failing Due to Missing ConfigMap Mount" }
func (l *PodMissingConfigMap2) Category() Category    { return CategoryWorkloads }
func (l *PodMissingConfigMap2) Difficulty() Difficulty { return DifficultyEasy }
func (l *PodMissingConfigMap2) EstimatedTime() int    { return 10 }
func (l *PodMissingConfigMap2) Tags() []string        { return []string{"configmap", "volumes", "mounts"} }

func (l *PodMissingConfigMap2) Description() string {
	return `A pod is failing because it references a ConfigMap that doesn't exist.
Create the missing ConfigMap.`
}

func (l *PodMissingConfigMap2) Hints() []string {
	return []string{
		"Check the pod volume configuration",
		"Verify the ConfigMap exists",
		"Create the ConfigMap with required data",
	}
}

func (l *PodMissingConfigMap2) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodMissingConfigMap2) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: config-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: config-app
  template:
    metadata:
      labels:
        app: config-app
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        volumeMounts:
        - name: config
          mountPath: /etc/config
      volumes:
      - name: config
        configMap:
          name: app-config-missing`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PodMissingConfigMap2) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deploy/config-app",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return err
	}
	if output != "1" {
		return fmt.Errorf("deployment not ready")
	}
	return nil
}

func (l *PodMissingConfigMap2) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check pod events", Command: "kubectl describe deploy config-app"},
		{Description: "Create ConfigMap", Command: "kubectl create configmap app-config-missing --from-literal=config.txt='hello'"},
	}
}
