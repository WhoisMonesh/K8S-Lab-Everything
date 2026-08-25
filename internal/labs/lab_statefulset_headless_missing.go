package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&StatefulsetHeadlessMissingLab{}) }

type StatefulsetHeadlessMissingLab struct{ BaseLab }

func (l *StatefulsetHeadlessMissingLab) ID() string          { return "statefulset_headless_missing" }
func (l *StatefulsetHeadlessMissingLab) Title() string        { return "StatefulSet Without Headless Service" }
func (l *StatefulsetHeadlessMissingLab) Category() Category   { return CategoryWorkloads }
func (l *StatefulsetHeadlessMissingLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *StatefulsetHeadlessMissingLab) EstimatedTime() int   { return 20 }
func (l *StatefulsetHeadlessMissingLab) Tags() []string {
	return []string{"statefulset", "headless", "service", "workloads"}
}
func (l *StatefulsetHeadlessMissingLab) Description() string {
	return `A StatefulSet named 'redis-cluster' cannot create pods — they stay
Pending because no headless service is defined.

A StatefulSet REQUIRES a headless Service (clusterIP: None) to provide
stable network identities (redis-0, redis-1, etc.).

Your task: Create the correct headless Service so the StatefulSet can
schedule its pods.`
}
func (l *StatefulsetHeadlessMissingLab) Hints() []string {
	return []string{
		"Check kubectl describe statefulset redis-cluster — look at events",
		"The StatefulSet spec has serviceName: redis-headless",
		"You need a Service named 'redis-headless' with clusterIP: None",
		"The selector must match the StatefulSet pod labels",
	}
}

func (l *StatefulsetHeadlessMissingLab) Break(ctx context.Context, kp string) error {
	sts := `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-cluster
  namespace: default
spec:
  serviceName: redis-headless
  replicas: 3
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
`
	return kubectlApply(ctx, kp, sts)
}

func (l *StatefulsetHeadlessMissingLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(5 * time.Second)
	return nil
}

func (l *StatefulsetHeadlessMissingLab) Verify(ctx context.Context, kp string) error {
	out, _ := kubectl(ctx, kp, "get", "svc", "redis-headless", "-o",
		"jsonpath={.spec.clusterIP}")
	if out != "None" {
		return fmt.Errorf("headless service missing or clusterIP is not None")
	}
	ready, _ := kubectl(ctx, kp, "get", "statefulset", "redis-cluster", "-o",
		"jsonpath={.status.readyReplicas}")
	if ready != "3" {
		return fmt.Errorf("statefulset not fully ready (ready: %s)", ready)
	}
	return nil
}

func (l *StatefulsetHeadlessMissingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Diagnose the failure", Command: "kubectl describe sts redis-cluster | tail -5", Notes: "Events show missing headless service"},
		{Description: "Create the headless Service", Command: `kubectl expose sts redis-cluster --name=redis-headless --clusterIP=None --port=6379`, Notes: "Must match the serviceName in StatefulSet spec"},
		{Description: "Verify pods come up", Command: "kubectl get pods -l app=redis -w", Notes: "Pods redis-0, redis-1, redis-2 created sequentially"},
		{Description: "Verify stable DNS", Command: "kubectl run dns-test --rm -it --image=busybox -- nslookup redis-0.redis-headless.default.svc.cluster.local", Notes: "Each pod gets a stable DNS name"},
	}
}
