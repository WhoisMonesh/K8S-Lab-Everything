package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&CMImmutableMigration{})
}

type CMImmutableMigration struct {
	BaseLab
}

func (l *CMImmutableMigration) ID() string             { return "cm_immutable_migration2" }
func (l *CMImmutableMigration) Title() string          { return "Immutable ConfigMap Migration" }
func (l *CMImmutableMigration) Category() Category     { return CategoryScheduling }
func (l *CMImmutableMigration) Difficulty() Difficulty { return DifficultyHard }
func (l *CMImmutableMigration) EstimatedTime() int     { return 20 }
func (l *CMImmutableMigration) Tags() []string {
	return []string{"configmap", "immutable", "migration"}
}

func (l *CMImmutableMigration) Description() string {
	return `A ConfigMap is set to immutable but needs to be updated.
Migrate to a new ConfigMap and update the pod references.`
}

func (l *CMImmutableMigration) Hints() []string {
	return []string{
		"Check if ConfigMap is immutable",
		"Create a new ConfigMap with updated data",
		"Update pod volume references",
	}
}

func (l *CMImmutableMigration) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CMImmutableMigration) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  config.yaml: |
    version: "1.0"
immutable: true
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
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
          name: app-config`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *CMImmutableMigration) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap", "app-config-v2",
		"-o", "jsonpath={.data.config\\.yaml}")
	if err != nil {
		return err
	}
	if output == "" {
		return fmt.Errorf("new ConfigMap not created")
	}
	return nil
}

func (l *CMImmutableMigration) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check ConfigMap", Command: "kubectl get configmap app-config -o yaml"},
		{Description: "Create new ConfigMap", Command: "kubectl create configmap app-config-v2 --from-literal=config.yaml='version: 2.0'"},
		{Description: "Update deployment", Command: "kubectl set env deployment/app CONFIG_MAP=app-config-v2"},
		{Description: "Update volume reference", Command: "kubectl edit deploy/app and change configMap name to app-config-v2"},
	}
}
