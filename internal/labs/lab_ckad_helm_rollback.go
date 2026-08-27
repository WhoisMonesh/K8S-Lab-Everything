package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADHelmRollbackLab{})
}

type CKADHelmRollbackLab struct {
	BaseLab
}

func (l *CKADHelmRollbackLab) ID() string             { return "ckad_helm_rollback" }
func (l *CKADHelmRollbackLab) Title() string          { return "Rollback Helm Release" }
func (l *CKADHelmRollbackLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADHelmRollbackLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADHelmRollbackLab) Cert() Cert             { return CertCKAD }
func (l *CKADHelmRollbackLab) DomainWeight() int      { return 20 }
func (l *CKADHelmRollbackLab) EstimatedTime() int     { return 15 }
func (l *CKADHelmRollbackLab) Tags() []string {
	return []string{"helm", "rollback", "recovery"}
}

func (l *CKADHelmRollbackLab) Description() string {
	return `A Helm release was upgraded to a broken version. The application is not
working correctly.

Your task: Rollback the Helm release to the previous revision.`
}

func (l *CKADHelmRollbackLab) Hints() []string {
	return []string{
		"Use helm rollback to revert to a previous revision",
		"Check helm history to see all revisions",
		"Rollback to revision 1 if it was the last working version",
	}
}

func (l *CKADHelmRollbackLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADHelmRollbackLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADHelmRollbackLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app.kubernetes.io/name=nginx",
		"-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}
	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pods not running (status: %s)", output)
	}
	return nil
}

func (l *CKADHelmRollbackLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check release history", Command: "helm history my-release"},
		{Description: "Rollback to previous revision", Command: "helm rollback my-release 1"},
		{Description: "Verify rollback", Command: "helm status my-release"},
	}
}
