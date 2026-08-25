package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodHostPathLab{})
}

type PodHostPathLab struct {
	BaseLab
}

func (l *PodHostPathLab) ID() string {
	return "pod_host_path_wrong"
}

func (l *PodHostPathLab) Title() string {
	return "Pod Failing Due to Wrong hostPath Mount"
}

func (l *PodHostPathLab) Category() Category {
	return CategoryStorage
}

func (l *PodHostPathLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *PodHostPathLab) Description() string {
	return `A pod 'data-processor' is failing because it mounts a hostPath that doesn't exist on the node.

Your task: Fix the hostPath to use a valid path on the node.`
}

func (l *PodHostPathLab) Hints() []string {
	return []string{
		"Check the pod status and events",
		"Look at the hostPath configuration",
		"The path /nonexistent/data doesn't exist on the node",
		"Use a valid path like /tmp/data or create the directory",
	}
}

func (l *PodHostPathLab) EstimatedTime() int {
	return 15
}

func (l *PodHostPathLab) Tags() []string {
	return []string{"hostpath", "storage", "volume", "node"}
}

func (l *PodHostPathLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodHostPathLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: data-processor
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.28
    command: ['sh', '-c', 'ls /data && echo "Data found" || echo "Data not found" && sleep 3600']
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    hostPath:
      path: /nonexistent/data
      type: DirectoryOrCreate
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}
	return nil
}

func (l *PodHostPathLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodHostPathLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "data-processor",
		"--", "ls", "/data")
	if err != nil {
		return fmt.Errorf("failed to exec into pod: %w", err)
	}
	if strings.Contains(output, "No such file") || strings.Contains(output, "not found") {
		return fmt.Errorf("hostPath mount is not working")
	}
	return nil
}

func (l *PodHostPathLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod status",
			Command:     "kubectl get pod data-processor",
			Notes:       "Pod may be running but the mount is wrong",
		},
		{
			Description: "Check the pod events",
			Command:     "kubectl describe pod data-processor | grep -A 10 Events",
			Notes:       "Look for volume mount warnings",
		},
		{
			Description: "Fix the hostPath",
			Command:     "kubectl delete pod data-processor",
			Notes:       "Delete the pod to recreate with correct hostPath",
		},
		{
			Description: "Create pod with valid hostPath",
			Command:     `kubectl run data-processor --image=busybox:1.28 --dry-run=client -o yaml > fixed-pod.yaml`,
			Notes:       "Edit the YAML to change hostPath from /nonexistent/data to /tmp/data",
		},
		{
			Description: "Apply the fixed pod",
			Command:     "kubectl apply -f fixed-pod.yaml",
			Notes:       "The pod should now mount /tmp/data successfully",
		},
		{
			Description: "Verify the mount works",
			Command:     "kubectl exec data-processor -- ls /data",
			Notes:       "Should list files from /tmp/data on the host",
		},
	}
}
