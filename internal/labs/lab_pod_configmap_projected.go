package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodConfigMapProjectedLab{})
}

type PodConfigMapProjectedLab struct {
	BaseLab
}

func (l *PodConfigMapProjectedLab) ID() string {
	return "pod_configmap_projected"
}

func (l *PodConfigMapProjectedLab) Title() string {
	return "Projected ConfigMap Volume Wrong"
}

func (l *PodConfigMapProjectedLab) Category() Category {
	return CategoryStorage
}

func (l *PodConfigMapProjectedLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodConfigMapProjectedLab) Description() string {
	return `A pod 'projected-app' uses a projected volume to combine a ConfigMap
and Secret, but the ConfigMap path in the projection is wrong. The app
cannot find its configuration at the expected location.

Your task: Fix the projected volume path for the ConfigMap.`
}

func (l *PodConfigMapProjectedLab) Hints() []string {
	return []string{
		"Check the projected volume configuration",
		"Each item in a projected volume has its own path",
		"Verify the path matches what the application expects",
	}
}

func (l *PodConfigMapProjectedLab) EstimatedTime() int {
	return 15
}

func (l *PodConfigMapProjectedLab) Tags() []string {
	return []string{"projected", "configmap", "volume", "storage"}
}

func (l *PodConfigMapProjectedLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodConfigMapProjectedLab) Break(ctx context.Context, kubeconfigPath string) error {
	configMap := `apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  config.yaml: |
    server:
      port: 8080
`
	if err := kubectlApply(ctx, kubeconfigPath, configMap); err != nil {
		return fmt.Errorf("creating ConfigMap: %w", err)
	}

	secret := `apiVersion: v1
kind: Secret
metadata:
  name: app-creds
  namespace: default
type: Opaque
data:
  password: cGFzc3dvcmQ=
`
	if err := kubectlApply(ctx, kubeconfigPath, secret); err != nil {
		return fmt.Errorf("creating Secret: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: projected-app
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'cat /etc/config/config.yaml && sleep 3600']
    volumeMounts:
    - name: projected-vol
      mountPath: /etc/config
  volumes:
  - name: projected-vol
    projected:
      sources:
      - configMap:
          name: app-config
          items:
          - key: config.yaml
            path: wrong-config/config.yaml
      - secret:
          name: app-creds
          items:
          - key: password
            path: creds/password
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodConfigMapProjectedLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodConfigMapProjectedLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "projected-app",
		"--", "cat", "/etc/config/config.yaml")
	if err != nil {
		return fmt.Errorf("cannot read config at expected path: %w", err)
	}

	if !strings.Contains(output, "port: 8080") {
		return fmt.Errorf("config not found at expected path")
	}

	return nil
}

func (l *PodConfigMapProjectedLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check projected volume",
			Command:     "kubectl get pod projected-app -o yaml | grep -A 20 projected",
			Notes:       "ConfigMap path is wrong-config/config.yaml instead of config.yaml",
		},
		{
			Description: "Fix the path",
			Command:     "kubectl edit pod projected-app",
			Notes:       "Change path from wrong-config/config.yaml to config.yaml",
		},
		{
			Description: "Verify config is accessible",
			Command:     "kubectl exec projected-app -- cat /etc/config/config.yaml",
			Notes:       "Should now show the config content",
		},
	}
}
