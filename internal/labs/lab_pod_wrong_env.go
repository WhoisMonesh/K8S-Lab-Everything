package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodWrongEnvLab{})
}

type PodWrongEnvLab struct {
	BaseLab
}

func (l *PodWrongEnvLab) ID() string {
	return "pod_wrong_env"
}

func (l *PodWrongEnvLab) Title() string {
	return "Pod Failing Due to Wrong Environment Variable Value"
}

func (l *PodWrongEnvLab) Category() Category {
	return CategoryWorkloads
}

func (l *PodWrongEnvLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *PodWrongEnvLab) Description() string {
	return `A pod 'cache-server' is crashing because the REDIS_HOST environment variable points to a wrong hostname.

Your task: Fix the environment variable to point to the correct Redis service.`
}

func (l *PodWrongEnvLab) Hints() []string {
	return []string{
		"Check the pod logs",
		"Look at the environment variables",
		"The REDIS_HOST value doesn't match the actual Redis service name",
		"Check what services exist in the namespace",
	}
}

func (l *PodWrongEnvLab) EstimatedTime() int {
	return 10
}

func (l *PodWrongEnvLab) Tags() []string {
	return []string{"environment", "env", "redis", "troubleshooting"}
}

func (l *PodWrongEnvLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodWrongEnvLab) Break(ctx context.Context, kubeconfigPath string) error {
	redisSvc := `apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: default
spec:
  selector:
    app: redis
  ports:
  - port: 6379
`
	if err := kubectlApply(ctx, kubeconfigPath, redisSvc); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: cache-server
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.28
    command: ['sh', '-c', 'echo "Connecting to $REDIS_HOST:$REDIS_PORT" && nc -z $REDIS_HOST $REDIS_PORT && echo "Connected!" || echo "Connection failed" && sleep 3600']
    env:
    - name: REDIS_HOST
      value: "redis-wrong-service"
    - name: REDIS_PORT
      value: "6379"
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}
	return nil
}

func (l *PodWrongEnvLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodWrongEnvLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "cache-server",
		"--", "printenv", "REDIS_HOST")
	if err != nil {
		return fmt.Errorf("failed to exec into pod: %w", err)
	}
	if strings.TrimSpace(output) != "redis" {
		return fmt.Errorf("REDIS_HOST is wrong (got: %s, expected: redis)", output)
	}
	return nil
}

func (l *PodWrongEnvLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod logs",
			Command:     "kubectl logs cache-server",
			Notes:       "Logs show connection failure to redis-wrong-service",
		},
		{
			Description: "Check available services",
			Command:     "kubectl get svc",
			Notes:       "The Redis service is named 'redis', not 'redis-wrong-service'",
		},
		{
			Description: "Fix the environment variable",
			Command:     "kubectl delete pod cache-server",
			Notes:       "Delete the pod to recreate with correct env",
		},
		{
			Description: "Create pod with correct REDIS_HOST",
			Command:     `kubectl run cache-server --image=busybox:1.28 --env="REDIS_HOST=redis" --env="REDIS_PORT=6379" -- sh -c 'nc -z $REDIS_HOST $REDIS_PORT && echo Connected || echo Failed; sleep 3600'`,
			Notes:       "Set REDIS_HOST to 'redis' which matches the service name",
		},
		{
			Description: "Verify connection",
			Command:     "kubectl logs cache-server",
			Notes:       "Should show 'Connected!' message",
		},
	}
}
