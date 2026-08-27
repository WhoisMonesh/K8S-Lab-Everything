package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&CMImmutableMigrationLab{}) }

type CMImmutableMigrationLab struct{ BaseLab }

func (l *CMImmutableMigrationLab) ID() string             { return "cm_immutable_migration" }
func (l *CMImmutableMigrationLab) Title() string          { return "Immutable ConfigMap Migration" }
func (l *CMImmutableMigrationLab) Category() Category     { return CategoryScheduling }
func (l *CMImmutableMigrationLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CMImmutableMigrationLab) EstimatedTime() int     { return 20 }
func (l *CMImmutableMigrationLab) Tags() []string {
	return []string{"configmap", "immutable", "mount", "scheduling"}
}
func (l *CMImmutableMigrationLab) Description() string {
	return `A deployment 'web-app' mounts a ConfigMap 'app-config' which was
accidentally created with immutable: true and contains outdated values.

Every attempt to kubectl edit the ConfigMap fails with: "the object has been
modified; please apply your changes to the default version of the object"

Your task: Create a NEW ConfigMap 'app-config-v2' with the corrected values
(env=production, debug=false), update the deployment to mount the new
ConfigMap, and verify the pod has the new values. Delete the old
ConfigMap after migration.`
}
func (l *CMImmutableMigrationLab) Hints() []string {
	return []string{
		"Immutable ConfigMaps cannot be edited — you must create a new one",
		"Use: kubectl create configmap app-config-v2 --from-literal=env=production --from-literal=debug=false",
		"Patch the deployment to mount app-config-v2 instead of app-config",
		"Delete the old app-config after confirming the new values work",
	}
}

func (l *CMImmutableMigrationLab) Break(ctx context.Context, kp string) error {
	cm := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
immutable: true
data:
  env: staging
  debug: "true"
`
	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web-app
  template:
    metadata:
      labels:
        app: web-app
    spec:
      containers:
      - name: web
        image: busybox:1.36
        command: ["sh","-c","cat /etc/config/env; cat /etc/config/debug; while true; do sleep 5; done"]
        volumeMounts:
        - name: config
          mountPath: /etc/config
      volumes:
      - name: config
        configMap:
          name: app-config
`
	kubectlApply(ctx, kp, cm)
	return kubectlApply(ctx, kp, deploy)
}

func (l *CMImmutableMigrationLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *CMImmutableMigrationLab) Verify(ctx context.Context, kp string) error {
	out, _ := kubectl(ctx, kp, "get", "cm", "app-config-v2", "-o",
		"jsonpath={.data.env}")
	if out != "production" {
		return fmt.Errorf("new ConfigMap has wrong env value: %s", out)
	}
	deployCM, _ := kubectl(ctx, kp, "get", "deploy", "web-app", "-o",
		"jsonpath={.spec.template.spec.volumes[0].configMap.name}")
	if deployCM != "app-config-v2" {
		return fmt.Errorf("deployment still mounts old ConfigMap: %s", deployCM)
	}
	podLogs, _ := kubectl(ctx, kp, "logs", "-l", "app=web-app", "--tail=2")
	if !strings.Contains(podLogs, "production") || strings.Contains(podLogs, "staging") {
		return fmt.Errorf("pod still shows old values: %s", podLogs)
	}
	return nil
}

func (l *CMImmutableMigrationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Verify old CM is immutable", Command: "kubectl get cm app-config -o yaml | grep immutable", Notes: "Shows immutable: true"},
		{Description: "Create new ConfigMap", Command: "kubectl create configmap app-config-v2 --from-literal=env=production --from-literal=debug=false", Notes: "New CM with correct values"},
		{Description: "Patch deployment to use new CM", Command: `kubectl patch deploy web-app -p '{"spec":{"template":{"spec":{"volumes":[{"name":"config","configMap":{"name":"app-config-v2"}}]}}}}'`, Notes: "Rolling update pulls new pod with new mount"},
		{Description: "Verify new values", Command: "kubectl logs -l app=web-app --tail=5", Notes: "Shows 'production' and 'false'"},
		{Description: "Clean up old CM", Command: "kubectl delete cm app-config", Notes: "Old immutable CM no longer referenced"},
	}
}
