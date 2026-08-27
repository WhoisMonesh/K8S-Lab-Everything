package labs

import (
	"context"
	"fmt"
	"time"
)

func init() {
	Register(&InitContainerFailLab{})
}

type InitContainerFailLab struct {
	BaseLab
}

func (l *InitContainerFailLab) ID() string {
	return "init_container_fail"
}

func (l *InitContainerFailLab) Title() string {
	return "App Blocked By Failing Init Container"
}

func (l *InitContainerFailLab) Category() Category {
	return CategoryWorkloads
}

func (l *InitContainerFailLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *InitContainerFailLab) Description() string {
	return `A deployment named 'migrator' never becomes ready. The main container is
fine, but the pod is stuck with status Init:0/1 forever.

The init container waits for a database service that does not exist in this
cluster before running schema migrations.

Your task: Get the deployment healthy. Either provide the dependency it waits
for or correct the init container logic - document which one you chose.`
}

func (l *InitContainerFailLab) Hints() []string {
	return []string{
		"Pod status Init:0/1 means the init container has not exited yet",
		"kubectl logs <pod> --all-containers shows what the init container prints",
		"The init loop does nslookup against db-svc.default.svc.cluster.local",
		"Check whether a Service named db-svc exists: kubectl get svc db-svc",
	}
}

func (l *InitContainerFailLab) EstimatedTime() int {
	return 20
}

func (l *InitContainerFailLab) Tags() []string {
	return []string{"init-containers", "dns", "dependency", "workloads", "troubleshooting"}
}

func (l *InitContainerFailLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *InitContainerFailLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: migrator
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: migrator
  template:
    metadata:
      labels:
        app: migrator
    spec:
      initContainers:
      - name: wait-for-db
        image: busybox:1.36
        command: ["sh", "-c", "until nslookup db-svc.default.svc.cluster.local; do echo waiting for db; sleep 3; done"]
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80
`

	if err := kubectlApply(ctx, kubeconfigPath, deployment); err != nil {
		return fmt.Errorf("creating broken deployment: %w", err)
	}
	return nil
}

func (l *InitContainerFailLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=migrator",
		"-o", "jsonpath={.items[*].status.phase}")
	for _, phase := range splitFields(output) {
		if phase == "Running" {
			return fmt.Errorf("expected pod stuck in init")
		}
	}
	return nil
}

func (l *InitContainerFailLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "migrator",
		"-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}
	if output != "1" {
		return fmt.Errorf("deployment not ready yet (ready replicas: %s, expected: 1)", output)
	}
	return nil
}

func (l *InitContainerFailLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Confirm where the pod is stuck",
			Command:     "kubectl get pods -l app=migrator",
			Notes:       "STATUS shows Init:0/1 - the app container never starts until init succeeds",
		},
		{
			Description: "Read the init container logs",
			Command:     "kubectl logs deploy/migrator --all-containers",
			Notes:       "It loops printing 'waiting for db' from a failing nslookup of db-svc",
		},
		{
			Description: "Verify the dependency is missing",
			Command:     "kubectl get svc db-svc",
			Notes:       "NotFound - there is no such service in this cluster",
		},
		{
			Description: "Fix option A: replace the init command with valid logic",
			Command:     "kubectl set image deploy/migrator wait-for-db=busybox:1.36 && kubectl patch deploy migrator --type='json' -p='[{\"op\":\"replace\",\"path\":\"/spec/template/spec/initContainers/0/command\",\"value\":[\"sh\",\"-c\",\"echo db already migrated; exit 0\"]}]'",
			Notes:       "In production you would point this at your real database host instead",
		},
		{
			Description: "Fix option B: create the missing service",
			Command:     "kubectl expose deploy/migrator --name=db-svc --port=5432",
			Notes:       "Acceptable for practice; nslookup then resolves and migrations 'run'",
		},
		{
			Description: "Watch the pod go Ready",
			Command:     "kubectl rollout status deploy/migrator && kubectl get pods -l app=migrator",
			Notes:       "STATUS should now show 1/1 Ready with no init delay",
		},
	}
}
