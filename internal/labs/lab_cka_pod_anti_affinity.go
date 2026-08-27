package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodAntiAffinityLab{})
}

type PodAntiAffinityLab struct {
	BaseLab
}

func (l *PodAntiAffinityLab) ID() string             { return "cka_pod_anti_affinity" }
func (l *PodAntiAffinityLab) Title() string          { return "Configure Pod Anti-Affinity Rules" }
func (l *PodAntiAffinityLab) Category() Category     { return CategoryWorkloadsScheduling }
func (l *PodAntiAffinityLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *PodAntiAffinityLab) EstimatedTime() int     { return 20 }
func (l *PodAntiAffinityLab) Tags() []string {
	return []string{"anti-affinity", "scheduling", "distribution"}
}
func (l *PodAntiAffinityLab) Cert() Cert        { return CertCKA }
func (l *PodAntiAffinityLab) DomainWeight() int { return 15 }

func (l *PodAntiAffinityLab) Description() string {
	return `A deployment has all replicas scheduled on the same node. Configure
pod anti-affinity to ensure replicas are spread across different nodes
for high availability.`
}

func (l *PodAntiAffinityLab) Hints() []string {
	return []string{
		"Add podAntiAffinity to the deployment spec",
		"Use requiredDuringSchedulingIgnoredDuringExecution",
		"Match labels to spread pods across nodes",
	}
}

func (l *PodAntiAffinityLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodAntiAffinityLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *PodAntiAffinityLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "spread-app",
		"-n", "affinity-ns", "-o",
		"jsonpath={.spec.template.spec.affinity.podAntiAffinity}")
	if err != nil {
		return err
	}
	if output == "" || output == "null" {
		return fmt.Errorf("pod anti-affinity not configured")
	}
	return nil
}

func (l *PodAntiAffinityLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check deployment", Command: "kubectl get deployment spread-app -n affinity-ns -o yaml"},
		{Description: "Add anti-affinity", Command: "kubectl patch deployment spread-app -n affinity-ns --type=strategic -p '{\"spec\":{\"template\":{\"spec\":{\"affinity\":{\"podAntiAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":[{\"labelSelector\":{\"matchExpressions\":[{\"key\":\"app\",\"operator\":\"In\",\"values\":[\"spread-app\"]}]},\"topologyKey\":\"kubernetes.io/hostname\"}]}}}}}}}'"},
		{Description: "Verify spread", Command: "kubectl get pods -n affinity-ns -o wide"},
	}
}
