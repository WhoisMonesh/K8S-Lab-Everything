package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodAffinitySelectorMismatchLab{})
}

type PodAffinitySelectorMismatchLab struct {
	BaseLab
}

func (l *PodAffinitySelectorMismatchLab) ID() string {
	return "pod_affinity_selector_mismatch"
}

func (l *PodAffinitySelectorMismatchLab) Title() string {
	return "Pod Affinity Label Selector Wrong"
}

func (l *PodAffinitySelectorMismatchLab) Category() Category {
	return CategoryScheduling
}

func (l *PodAffinitySelectorMismatchLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodAffinitySelectorMismatchLab) Description() string {
	return `A Deployment 'cache-proxy' has a pod anti-affinity rule that prevents
it from being scheduled with pods having label app=database. However,
the selector uses app=db (wrong label), so the anti-affinity has no
effect and pods can land on the same node as the database.

Your task: Fix the pod anti-affinity selector to use the correct label.`
}

func (l *PodAffinitySelectorMismatchLab) Hints() []string {
	return []string{
		"Check the deployment's affinity rules",
		"Compare the affinity selector with actual pod labels",
		"The selector label doesn't match the database pods",
	}
}

func (l *PodAffinitySelectorMismatchLab) EstimatedTime() int {
	return 15
}

func (l *PodAffinitySelectorMismatchLab) Tags() []string {
	return []string{"affinity", "anti-affinity", "scheduling", "labels"}
}

func (l *PodAffinitySelectorMismatchLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodAffinitySelectorMismatchLab) Break(ctx context.Context, kubeconfigPath string) error {
	database := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: database
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: database
  template:
    metadata:
      labels:
        app: database
        tier: backend
    spec:
      containers:
      - name: db
        image: busybox:1.36
        command: ['sh', '-c', 'sleep 3600']
`
	if err := kubectlApply(ctx, kubeconfigPath, database); err != nil {
		return fmt.Errorf("creating database: %w", err)
	}

	cacheProxy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: cache-proxy
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cache-proxy
  template:
    metadata:
      labels:
        app: cache-proxy
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchExpressions:
              - key: app
                operator: In
                values:
                - db
            topologyKey: kubernetes.io/hostname
      containers:
      - name: cache
        image: nginx:alpine
        ports:
        - containerPort: 80
`
	if err := kubectlApply(ctx, kubeconfigPath, cacheProxy); err != nil {
		return fmt.Errorf("creating cache-proxy: %w", err)
	}

	return nil
}

func (l *PodAffinitySelectorMismatchLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodAffinitySelectorMismatchLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "cache-proxy",
		"-o", "jsonpath={.spec.template.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[0].labelSelector.matchExpressions[0].values[0]}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}

	if strings.TrimSpace(output) == "db" {
		return fmt.Errorf("anti-affinity selector still uses 'db' instead of 'database'")
	}

	return nil
}

func (l *PodAffinitySelectorMismatchLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check anti-affinity rules",
			Command:     "kubectl get deployment cache-proxy -o yaml | grep -A 10 podAntiAffinity",
			Notes:       "Selector uses app=db which doesn't match database pods",
		},
		{
			Description: "Check database pod labels",
			Command:     "kubectl get pods --show-labels | grep database",
			Notes:       "Database pods have app=database, not app=db",
		},
		{
			Description: "Fix anti-affinity selector",
			Command:     "kubectl edit deployment cache-proxy",
			Notes:       "Change 'db' to 'database' in the anti-affinity selector values",
		},
		{
			Description: "Verify fix",
			Command:     "kubectl get deployment cache-proxy -o yaml | grep -A 10 podAntiAffinity",
			Notes:       "Should now use app=database",
		},
	}
}
