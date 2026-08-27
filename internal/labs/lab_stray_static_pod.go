package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&StrayStaticPodLab{}) }

type StrayStaticPodLab struct{ BaseLab }

func (l *StrayStaticPodLab) ID() string             { return "stray_static_pod" }
func (l *StrayStaticPodLab) Title() string          { return "Stray Static Pod Consuming Resources" }
func (l *StrayStaticPodLab) Category() Category     { return CategoryControlPlane }
func (l *StrayStaticPodLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *StrayStaticPodLab) EstimatedTime() int     { return 15 }
func (l *StrayStaticPodLab) Tags() []string {
	return []string{"static-pod", "control-plane", "kubelet"}
}
func (l *StrayStaticPodLab) Description() string {
	return `A stray static pod named 'orphan-nginx' is running on the control-
plane node, consuming resources. It was created by a manifest file
left in /etc/kubernetes/manifests/ (or equivalent on the node).

You cannot delete this pod with kubectl — it's a static pod managed
by kubelet. It keeps respawning.

Your task: Find and remove the static pod manifest file so kubelet
stops managing it. You'll need to exec into the node (or access the
host filesystem) to remove the YAML file.`
}
func (l *StrayStaticPodLab) Hints() []string {
	return []string{
		"Static pods show up with a node suffix: orphan-nginx-<node-name>",
		"They're managed by kubelet from manifest files in /etc/kubernetes/manifests/",
		"Use: kubectl get pod orphan-nginx-<node> -n kube-system to find it",
		"Remove the manifest file: docker exec <node> rm /etc/kubernetes/manifests/orphan-nginx.yaml",
		"kubelet will stop the pod within seconds of deleting the manifest",
	}
}

func (l *StrayStaticPodLab) Break(ctx context.Context, kp string) error {
	// Find the control-plane node
	node, err := getControlPlaneNode(ctx, kp)
	if err != nil {
		return fmt.Errorf("cannot find control-plane node: %w", err)
	}
	node = strings.TrimSpace(node)

	// Create the static pod manifest via docker exec
	manifest := `apiVersion: v1
kind: Pod
metadata:
  name: orphan-nginx
  namespace: kube-system
spec:
  containers:
  - name: nginx
    image: nginx:1.27-alpine
    ports:
    - containerPort: 80
    resources:
      requests:
        cpu: 500m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 256Mi
`

	// Write the manifest via docker
	writeCmd := fmt.Sprintf(`cat <<'MANIFEST' > /etc/kubernetes/manifests/orphan-nginx.yaml
%s
MANIFEST`, manifest)
	_, err = dockerExec(ctx, node, "sh", "-c", writeCmd)
	return err
}

func (l *StrayStaticPodLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *StrayStaticPodLab) Verify(ctx context.Context, kp string) error {
	// Check the static pod no longer exists
	node, _ := getControlPlaneNode(ctx, kp)
	node = strings.TrimSpace(node)
	podName := "orphan-nginx-" + node

	out, err := kubectl(ctx, kp, "get", "pod", podName, "-n", "kube-system", "-o", "name")
	if err == nil && strings.Contains(out, podName) {
		phase, _ := kubectl(ctx, kp, "get", "pod", podName, "-n", "kube-system", "-o", "jsonpath={.status.phase}")
		if phase != "" {
			return fmt.Errorf("static pod %s still exists (phase: %s)", podName, phase)
		}
	}
	return nil
}

func (l *StrayStaticPodLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Find the stray static pod", Command: "kubectl get pods -n kube-system -o wide | grep orphan", Notes: "Shows orphan-nginx-<node>"},
		{Description: "Identify the node", Command: "kubectl get pods -n kube-system orphan-nginx-<node> -o jsonpath='{.spec.nodeName}'", Notes: "Determines which node hosts the static pod"},
		{Description: "Remove the manifest from the node", Command: "docker exec <control-plane-node> rm /etc/kubernetes/manifests/orphan-nginx.yaml", Notes: "kubelet stops the pod within seconds"},
		{Description: "Verify removal", Command: "kubectl get pods -n kube-system | grep orphan", Notes: "Pod no longer listed"},
	}
}
