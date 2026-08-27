package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&LivenessProbeFlapLab{})
}

type LivenessProbeFlapLab struct {
	BaseLab
}

func (l *LivenessProbeFlapLab) ID() string {
	return "liveness_probe_flap"
}

func (l *LivenessProbeFlapLab) Title() string {
	return "Misconfigured Liveness Probe Restarting Pods"
}

func (l *LivenessProbeFlapLab) Category() Category {
	return CategoryWorkloads
}

func (l *LivenessProbeFlapLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *LivenessProbeFlapLab) Description() string {
	return `A web frontend deployment named 'shop' appears healthy at first but every
pod gets restarted roughly every minute. RESTARTS in kubectl get pods keeps
climbing and users report intermittent failures.

Your task: Find out why Kubernetes keeps killing the containers and fix it so
the pods run stably without disabling health checks entirely.`
}

func (l *LivenessProbeFlapLab) Hints() []string {
	return []string{
		"kubectl describe pod shows events like 'Liveness probe failed' and 'Restarting container'",
		"Compare the probe port/path against the ports the container actually serves",
		"kubectl logs will show nginx serving on port 80 only",
		"A liveness probe must point at an endpoint the app actually answers",
	}
}

func (l *LivenessProbeFlapLab) EstimatedTime() int {
	return 20
}

func (l *LivenessProbeFlapLab) Tags() []string {
	return []string{"probes", "liveness", "healthcheck", "restarts", "troubleshooting"}
}

func (l *LivenessProbeFlapLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *LivenessProbeFlapLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: shop
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: shop
  template:
    metadata:
      labels:
        app: shop
    spec:
      containers:
      - name: frontend
        image: nginx:alpine
        ports:
        - containerPort: 80
        livenessProbe:
          httpGet:
            path: /
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 10
`

	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating broken deployment: %w", err)
	}
	return nil
}

func (l *LivenessProbeFlapLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(30 * time.Second)
	return nil
}

func (l *LivenessProbeFlapLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "shop",
		"-o", "jsonpath={.spec.template.spec.containers[0].livenessProbe.httpGet.port}")
	if err != nil {
		return fmt.Errorf("failed to check probe config: %w", err)
	}
	if output == "8080" {
		return fmt.Errorf("liveness probe still points at port 8080 which nothing listens on")
	}

	ready, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "shop",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}
	if ready != "2" {
		return fmt.Errorf("deployment not fully ready yet (ready replicas: %s, expected: 2)", ready)
	}

	restarts, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=shop",
		"-o", "jsonpath={.items[*].status.containerStatuses[*].restartCount}")
	if err != nil {
		return fmt.Errorf("failed to check restarts: %w", err)
	}
	for _, count := range splitFields(restarts) {
		if count != "0" {
			return fmt.Errorf("pods are still restarting - wait for fresh pods or delete the old ones")
		}
	}
	return nil
}

func (l *LivenessProbeFlapLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Watch the restart counter climb",
			Command:     "kubectl get pods -l app=shop -w",
			Notes:       "RESTARTS increases about once a minute per pod; press Ctrl+C when convinced",
		},
		{
			Description: "Find the kill reason in events",
			Command:     "kubectl describe pod -l app=shop | grep -B 2 -A 4 'Liveness probe failed'",
			Notes:       "Events show connection refused on port 8080 followed by container kills",
		},
		{
			Description: "Confirm what the container serves",
			Command:     "kubectl get deploy shop -o jsonpath='{.spec.template.spec.containers[0].ports}'",
			Notes:       "nginx listens on port 80 only; the liveness probe targets 8080",
		},
		{
			Description: "Point the liveness probe at the real port",
			Command:     "kubectl patch deploy shop --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/livenessProbe/httpGet/port\",\"value\":80}]'",
			Notes:       "kubectl edit deploy shop also works - change livenessProbe.httpGet.port to 80",
		},
		{
			Description: "Verify pods stabilize",
			Command:     "kubectl rollout status deploy/shop && kubectl get pods -l app=shop",
			Notes:       "New pods should hold RESTARTS at 0; delete any old flapping pods if needed",
		},
	}
}
