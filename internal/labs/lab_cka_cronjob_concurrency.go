package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&CronJobConcurrencyLab{})
}

type CronJobConcurrencyLab struct {
	BaseLab
}

func (l *CronJobConcurrencyLab) ID() string             { return "cka_cronjob_concurrency" }
func (l *CronJobConcurrencyLab) Title() string          { return "Configure CronJob Concurrency Policies" }
func (l *CronJobConcurrencyLab) Category() Category     { return CategoryWorkloadsScheduling }
func (l *CronJobConcurrencyLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CronJobConcurrencyLab) EstimatedTime() int     { return 15 }
func (l *CronJobConcurrencyLab) Tags() []string {
	return []string{"cronjob", "concurrency", "scheduling"}
}
func (l *CronJobConcurrencyLab) Cert() Cert        { return CertCKA }
func (l *CronJobConcurrencyLab) DomainWeight() int { return 15 }

func (l *CronJobConcurrencyLab) Description() string {
	return `A CronJob is using Allow concurrent policy causing resource contention.
Change it to Forbid policy to prevent overlapping runs and add a
startingDeadlineSeconds to handle missed schedules.`
}

func (l *CronJobConcurrencyLab) Hints() []string {
	return []string{
		"Check the CronJob concurrencyPolicy",
		"Change from Allow to Forbid",
		"Add startingDeadlineSeconds for missed schedule handling",
	}
}

func (l *CronJobConcurrencyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CronJobConcurrencyLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CronJobConcurrencyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "cronjob", "batch-job",
		"-n", "cronjob-ns", "-o", "jsonpath={.spec.concurrencyPolicy}")
	if err != nil {
		return err
	}
	if output == "Allow" {
		return fmt.Errorf("concurrency policy still set to Allow")
	}
	return nil
}

func (l *CronJobConcurrencyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check CronJob config", Command: "kubectl get cronjob batch-job -n cronjob-ns -o yaml"},
		{Description: "Patch concurrency policy", Command: "kubectl patch cronjob batch-job -n cronjob-ns -p '{\"spec\":{\"concurrencyPolicy\":\"Forbid\",\"startingDeadlineSeconds\":200}}'"},
		{Description: "Verify", Command: "kubectl get cronjob batch-job -n cronjob-ns -o yaml"},
	}
}
