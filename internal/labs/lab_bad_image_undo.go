package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&BadImageUndoLab{}) }

type BadImageUndoLab struct{ BaseLab }

func (l *BadImageUndoLab) ID() string             { return "bad_image_undo" }
func (l *BadImageUndoLab) Title() string          { return "Roll Back a Bad Image Update" }
func (l *BadImageUndoLab) Category() Category     { return CategoryWorkloads }
func (l *BadImageUndoLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *BadImageUndoLab) EstimatedTime() int     { return 15 }
func (l *BadImageUndoLab) Tags() []string {
	return []string{"deployment", "rollback", "rollout", "workloads"}
}
func (l *BadImageUndoLab) Description() string {
	return `A deployment named 'payments' was patched from nginx:1.26 to a typo
image 'ngnix:1.26'. The rollout is stuck on ImagePullBackOff.

Your task: Roll the deployment back to the previous working revision using
kubectl rollout undo (do not edit the YAML manually).`
}
func (l *BadImageUndoLab) Hints() []string {
	return []string{
		"Check kubectl rollout history deploy/payments",
		"Use kubectl rollout undo to revert to the last good revision",
		"After undo, kubectl rollout status confirms recovery",
	}
}

func (l *BadImageUndoLab) Break(ctx context.Context, kp string) error {
	// Apply v1
	v1 := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: payments
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: payments
  template:
    metadata:
      labels:
        app: payments
    spec:
      containers:
      - name: web
        image: nginx:1.26-alpine
`
	if err := kubectlApply(ctx, kp, v1); err != nil {
		return err
	}
	time.Sleep(5 * time.Second)
	// Patch to broken image
	_, err := kubectl(ctx, kp, "set", "image", "deploy/payments", "web=ngnix:1.26", "--record")
	return err
}

func (l *BadImageUndoLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *BadImageUndoLab) Verify(ctx context.Context, kp string) error {
	out, _ := kubectl(ctx, kp, "get", "deploy", "payments", "-o",
		"jsonpath={.spec.template.spec.containers[0].image}")
	if strings.Contains(out, "ngnix") {
		return fmt.Errorf("deployment still uses broken image ngnix:1.26")
	}
	ready, _ := kubectl(ctx, kp, "get", "deploy", "payments", "-o", "jsonpath={.status.readyReplicas}")
	if ready != "2" {
		return fmt.Errorf("not all replicas ready (ready: %s)", ready)
	}
	return nil
}

func (l *BadImageUndoLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "See the broken rollout", Command: "kubectl rollout history deploy/payments", Notes: "Revision 2 has the typo image"},
		{Description: "Check current image", Command: "kubectl get deploy payments -o jsonpath='{.spec.template.spec.containers[0].image}'", Notes: "Shows ngnix:1.26 — the typo"},
		{Description: "Roll back", Command: "kubectl rollout undo deploy/payments", Notes: "Reverts to previous revision (nginx:1.26-alpine)"},
		{Description: "Confirm", Command: "kubectl rollout status deploy/payments && kubectl get pods -l app=payments", Notes: "All pods Running with correct image"},
	}
}
