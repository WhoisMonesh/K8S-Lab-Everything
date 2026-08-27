package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&SchedulerConfigLab{})
}

type SchedulerConfigLab struct {
	BaseLab
}

func (l *SchedulerConfigLab) ID() string             { return "cka_scheduler_config" }
func (l *SchedulerConfigLab) Title() string          { return "Configure Scheduler with Multiple Profiles" }
func (l *SchedulerConfigLab) Category() Category     { return CategoryClusterArchitecture }
func (l *SchedulerConfigLab) Difficulty() Difficulty { return DifficultyHard }
func (l *SchedulerConfigLab) EstimatedTime() int     { return 30 }
func (l *SchedulerConfigLab) Tags() []string {
	return []string{"scheduler", "configuration", "profiles"}
}
func (l *SchedulerConfigLab) Cert() Cert        { return CertCKA }
func (l *SchedulerConfigLab) DomainWeight() int { return 25 }

func (l *SchedulerConfigLab) Description() string {
	return `Configure the Kubernetes scheduler with two scheduling profiles:
one for default pods and one for high-priority batch workloads. The batch
profile should use a different plugin configuration.`
}

func (l *SchedulerConfigLab) Hints() []string {
	return []string{
		"Create a KubeSchedulerConfiguration file",
		"Define multiple scheduling profiles",
		"Reference the config in the scheduler manifest",
	}
}

func (l *SchedulerConfigLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *SchedulerConfigLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *SchedulerConfigLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap", "kube-scheduler",
		"-n", "kube-system", "-o", "yaml")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "schedulerConfiguration") {
		return fmt.Errorf("scheduler configuration not found")
	}
	return nil
}

func (l *SchedulerConfigLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create scheduler config", Command: "cat <<EOF | sudo tee /etc/kubernetes/scheduler-config.yaml\napiVersion: kubescheduler.config.k8s.io/v1\nkind: KubeSchedulerConfiguration\nprofiles:\n- schedulerName: default-scheduler\n  plugins:\n    score:\n      enabled:\n      - name: NodeResourcesFit\n- schedulerName: batch-scheduler\n  plugins:\n    score:\n      enabled:\n      - name: NodeResourcesFit\n        weight: 2\nEOF"},
		{Description: "Add config to scheduler manifest", Command: "sudo sed -i '/--profiling/a\\\\    - --config=/etc/kubernetes/scheduler-config.yaml' /etc/kubernetes/manifests/kube-scheduler.yaml"},
		{Description: "Verify scheduler restart", Command: "kubectl get pods -n kube-system -l component=kube-scheduler"},
	}
}
