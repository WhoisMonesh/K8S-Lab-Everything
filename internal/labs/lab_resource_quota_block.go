package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ResourceQuotaBlockLab{})
}

type ResourceQuotaBlockLab struct {
	BaseLab
}

func (l *ResourceQuotaBlockLab) ID() string {
	return "resource_quota_block"
}

func (l *ResourceQuotaBlockLab) Title() string {
	return "Replicas Rejected By ResourceQuota"
}

func (l *ResourceQuotaBlockLab) Category() Category {
	return CategoryWorkloads
}

func (l *ResourceQuotaBlockLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ResourceQuotaBlockLab) Description() string {
	return `In the 'quota-lab' namespace a deployment named 'inventory' wants 2
replicas, but the ReplicaSet keeps logging FailedCreate events: the namespace
ResourceQuota is too small for the pods' resource requests.

Your task: Get both replicas Running while KEEPING the ResourceQuota in place.
You may either raise the quota or tune the pod resources so they fit - both
are valid production answers.`
}

func (l *ResourceQuotaBlockLab) Hints() []string {
	return []string{
		"kubectl describe rs -n quota-lab shows the exact quota rejection message",
		"Compare the quota hard limits with the pods' requests and limits",
		"Each inventory pod asks for 500m CPU / 256Mi memory but the quota caps far less",
		"kubectl edit resourcequota -n quota-lab or kubectl set resources deploy/inventory -n quota-lab",
	}
}

func (l *ResourceQuotaBlockLab) EstimatedTime() int {
	return 20
}

func (l *ResourceQuotaBlockLab) Tags() []string {
	return []string{"resourcequota", "admission", "capacity-planning", "workloads", "troubleshooting"}
}

func (l *ResourceQuotaBlockLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ResourceQuotaBlockLab) Break(ctx context.Context, kubeconfigPath string) error {
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: quota-lab
`

	quota := `apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-quota
  namespace: quota-lab
spec:
  hard:
    pods: "10"
    requests.cpu: "300m"
    requests.memory: "192Mi"
    limits.cpu: "600m"
    limits.memory: "384Mi"
`

	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: inventory
  namespace: quota-lab
spec:
  replicas: 2
  selector:
    matchLabels:
      app: inventory
  template:
    metadata:
      labels:
        app: inventory
    spec:
      containers:
      - name: api
        image: nginx:alpine
        ports:
        - containerPort: 80
        resources:
          requests:
            cpu: "500m"
            memory: "256Mi"
          limits:
            cpu: "500m"
            memory: "256Mi"
`

	for _, manifest := range []string{namespace, quota, deployment} {
		if err := kubectlApply(ctx, kubeconfigPath, manifest); err != nil {
			return fmt.Errorf("applying lab resources: %w", err)
		}
	}
	return nil
}

func (l *ResourceQuotaBlockLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	output, _ := kubectl(ctx, kubeconfigPath, "get", "rs", "-n", "quota-lab",
		"-l", "app=inventory", "-o", "jsonpath={.items[*].status.replicas}")
	if output == "" || output == "0" {
		return nil
	}
	return fmt.Errorf("expected all replicas to be rejected by quota")
}

func (l *ResourceQuotaBlockLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "resourcequota", "-n", "quota-lab")
	if err != nil {
		return fmt.Errorf("the ResourceQuota must remain in place: %w", err)
	}
	if containsAny(output, "NotFound") || output == "" {
		return fmt.Errorf("team-quota was deleted - fix it by resizing quota or pod resources instead")
	}

	ready, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "inventory",
		"-n", "quota-lab", "-o", "jsonpath={.status.readyReplicas}")
	if err != nil {
		return fmt.Errorf("failed to check deployment: %w", err)
	}
	if ready != "2" {
		return fmt.Errorf("deployment not fully ready yet (ready replicas: %s, expected: 2)", ready)
	}
	return nil
}

func (l *ResourceQuotaBlockLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "See the admission rejection",
			Command:     "kubectl describe rs -n quota-lab -l app=inventory | grep -A 3 Events",
			Notes:       "'exceeded quota: team-quota' names the exceeded dimension (requests.cpu)",
		},
		{
			Description: "Review quota vs pod requirements",
			Command:     "kubectl get resourcequota team-quota -n quota-lab && kubectl get deploy inventory -n quota-lab -o jsonpath='{.spec.template.spec.containers[0].resources}'",
			Notes:       "Pods request 500m/256Mi each; quota allows only 300m/192Mi total",
		},
		{
			Description: "Fix option A: raise the quota to fit two pods",
			Command:     "kubectl patch resourcequota team-quota -n quota-lab -p '{\"spec\":{\"hard\":{\"requests.cpu\":\"1200m\",\"requests.memory\":\"768Mi\",\"limits.cpu\":\"1200m\",\"limits.memory\":\"768Mi\"}}}'",
			Notes:       "Keep pods count as-is; only compute dimensions needed changing",
		},
		{
			Description: "Fix option B: shrink pod resources into the quota",
			Command:     "kubectl set resources deploy/inventory -n quota-lab --requests=cpu=100m,memory=64Mi --limits=cpu=200m,memory=128Mi",
			Notes:       "nginx idles happily at these levels; total now fits inside the existing quota",
		},
		{
			Description: "Confirm both replicas are Ready with quota intact",
			Command:     "kubectl rollout status deploy/inventory -n quota-lab && kubectl get resourcequota team-quota -n quota-lab",
			Notes:       "Used column should now reflect the two running pods",
		},
	}
}
