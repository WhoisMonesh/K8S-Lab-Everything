package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&ClusterAutoscalerLab{})
}

type ClusterAutoscalerLab struct {
	BaseLab
}

func (l *ClusterAutoscalerLab) ID() string { return "cka_cluster_autoscaler" }
func (l *ClusterAutoscalerLab) Title() string {
	return "Configure Cluster Autoscaler"
}
func (l *ClusterAutoscalerLab) Category() Category     { return CategoryClusterArchitecture }
func (l *ClusterAutoscalerLab) Difficulty() Difficulty { return DifficultyHard }
func (l *ClusterAutoscalerLab) EstimatedTime() int     { return 30 }
func (l *ClusterAutoscalerLab) Tags() []string {
	return []string{"autoscaler", "cluster", "scaling"}
}
func (l *ClusterAutoscalerLab) Cert() Cert        { return CertCKA }
func (l *ClusterAutoscalerLab) DomainWeight() int { return 25 }

func (l *ClusterAutoscalerLab) Description() string {
	return `The cluster autoscaler deployment is not properly configured. Fix the
deployment to point to the correct cluster name and ensure it has the proper
RBAC permissions to scale node groups.`
}

func (l *ClusterAutoscalerLab) Hints() []string {
	return []string{
		"Check the autoscaler deployment environment variables",
		"Verify the ClusterAutoscaler RBAC resources",
		"Ensure --skip-nodes-with-local-storage is configured",
	}
}

func (l *ClusterAutoscalerLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ClusterAutoscalerLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *ClusterAutoscalerLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "cluster-autoscaler",
		"-n", "kube-system", "-o", "jsonpath={.spec.template.spec.containers[0].env}")
	if err != nil {
		return err
	}
	if strings.Contains(output, "WRONG_CLUSTER") {
		return fmt.Errorf("autoscaler still has wrong cluster name")
	}
	return nil
}

func (l *ClusterAutoscalerLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check autoscaler config", Command: "kubectl get deployment cluster-autoscaler -n kube-system -o yaml"},
		{Description: "Fix cluster name", Command: "kubectl set env deployment/cluster-autoscaler -n kube-system CLUSTER_NAME=<correct-name>"},
		{Description: "Verify rollout", Command: "kubectl rollout status deployment/cluster-autoscaler -n kube-system"},
	}
}
