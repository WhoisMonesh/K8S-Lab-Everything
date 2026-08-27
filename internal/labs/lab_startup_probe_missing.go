package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&StartupProbeMissingLab{}) }

type StartupProbeMissingLab struct{ BaseLab }

func (l *StartupProbeMissingLab) ID() string             { return "startup_probe_missing" }
func (l *StartupProbeMissingLab) Title() string          { return "Liveness Probe Kills Slow-Starting App" }
func (l *StartupProbeMissingLab) Category() Category     { return CategoryWorkloads }
func (l *StartupProbeMissingLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *StartupProbeMissingLab) EstimatedTime() int     { return 20 }
func (l *StartupProbeMissingLab) Tags() []string {
	return []string{"probes", "startup", "liveness", "workloads"}
}
func (l *StartupProbeMissingLab) Description() string {
	return `A deployment named 'api-server' keeps restarting. The container takes
~30s to become ready on each start, but the liveness probe kicks in after
5 seconds and kills it every time — classic startup race.

Your task: Add a startupProbe that gives the app enough time to initialize,
so the liveness probe only runs after startup succeeds.`
}
func (l *StartupProbeMissingLab) Hints() []string {
	return []string{
		"Check kubectl describe pod — liveness failures happen within seconds of start",
		"The app writes a ready-file after ~20s: cat /tmp/ready",
		"Use a startupProbe with failureThreshold * periodSeconds > warmup time",
		"Once startupProbe passes, liveness takes over normally",
	}
}

func (l *StartupProbeMissingLab) Break(ctx context.Context, kp string) error {
	if _, err := kubectl(ctx, kp, "create", "ns", "probe-lab"); err != nil {
		return err
	}
	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
  namespace: probe-lab
spec:
  replicas: 2
  selector:
    matchLabels:
      app: api-server
  template:
    metadata:
      labels:
        app: api-server
    spec:
      containers:
      - name: api
        image: busybox:1.36
        command: ["sh","-c","touch /tmp/starting; sleep 25; touch /tmp/ready; echo ready; while true; do sleep 5; done"]
        livenessProbe:
          exec:
            command: ["sh","-c","test -f /tmp/ready"]
          initialDelaySeconds: 5
          periodSeconds: 5
          failureThreshold: 3
`
	return kubectlApply(ctx, kp, deploy)
}

func (l *StartupProbeMissingLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(45 * time.Second)
	return nil
}

func (l *StartupProbeMissingLab) Verify(ctx context.Context, kp string) error {
	out, err := kubectl(ctx, kp, "get", "deploy", "api-server", "-n", "probe-lab",
		"-o", "jsonpath={.spec.template.spec.containers[0].startupProbe}")
	if err != nil {
		return err
	}
	if strings.TrimSpace(out) == "" {
		return fmt.Errorf("no startupProbe found — liveness will keep killing during warmup")
	}
	ready, _ := kubectl(ctx, kp, "get", "deploy", "api-server", "-n", "probe-lab",
		"-o", "jsonpath={.status.readyReplicas}")
	if ready != "2" {
		return fmt.Errorf("not all replicas ready yet (ready: %s)", ready)
	}
	return nil
}

func (l *StartupProbeMissingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Observe the restart loop", Command: "kubectl get pods -n probe-lab -w", Notes: "Pods restart every ~15s as liveness kills them during startup"},
		{Description: "Diagnose from events", Command: "kubectl describe pod -n probe-lab | grep -A5 Liveness", Notes: "Liveness exec test fails because /tmp/ready isn't written yet"},
		{Description: "Add a startupProbe to give time", Command: `kubectl patch deploy api-server -n probe-lab --type='json' -p='[{"op":"add","path":"/spec/template/spec/containers/0/startupProbe","value":{"exec":{"command":["sh","-c","test -f /tmp/ready"]},"initialDelaySeconds":5,"periodSeconds":5,"failureThreshold":10}}]'`, Notes: "failureThreshold(10)*periodSeconds(5)=50s of grace before liveness activates"},
		{Description: "Verify rollout", Command: "kubectl rollout status deploy/api-server -n probe-lab", Notes: "New pods will survive startup and stay Running"},
	}
}
