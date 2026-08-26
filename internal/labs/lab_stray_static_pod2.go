package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&StrayStaticPod{})
}

type StrayStaticPod struct {
	BaseLab
}

func (l *StrayStaticPod) ID() string            { return "stray_static_pod2" }
func (l *StrayStaticPod) Title() string         { return "Stray Static Pod Consuming Resources" }
func (l *StrayStaticPod) Category() Category    { return CategoryControlPlane }
func (l *StrayStaticPod) Difficulty() Difficulty { return DifficultyMedium }
func (l *StrayStaticPod) EstimatedTime() int    { return 15 }
func (l *StrayStaticPod) Tags() []string        { return []string{"static-pods", "kubelet", "resources"} }

func (l *StrayStaticPod) Description() string {
	return `A stray static pod is consuming resources on a node.
Identify and remove the static pod manifest.`
}

func (l *StrayStaticPod) Hints() []string {
	return []string{
		"Check for static pods on the node",
		"Look at /etc/kubernetes/manifests/",
		"Remove the unwanted manifest",
	}
}

func (l *StrayStaticPod) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *StrayStaticPod) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Pod
metadata:
  name: stray-nginx
  namespace: kube-system
spec:
  containers:
  - name: nginx
    image: nginx:alpine
    ports:
    - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *StrayStaticPod) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return err
	}
	if containsAny(output, "stray-nginx") {
		return fmt.Errorf("stray static pod still exists")
	}
	return nil
}

func (l *StrayStaticPod) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check static pods", Command: "kubectl get pods -n kube-system"},
		{Description: "Identify static pod", Command: "kubectl get pods -n kube-system -o wide | grep <node-name>"},
		{Description: "Remove manifest", Command: "ssh <node> sudo rm /etc/kubernetes/manifests/stray-nginx.yaml"},
	}
}
