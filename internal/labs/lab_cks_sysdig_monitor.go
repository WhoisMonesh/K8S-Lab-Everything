package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSSysdigMonitorLab{})
}

type CKSSysdigMonitorLab struct {
	BaseLab
}

func (l *CKSSysdigMonitorLab) ID() string             { return "cks_sysdig_monitor" }
func (l *CKSSysdigMonitorLab) Title() string          { return "Use Sysdig for Monitoring" }
func (l *CKSSysdigMonitorLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSSysdigMonitorLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSSysdigMonitorLab) EstimatedTime() int     { return 25 }
func (l *CKSSysdigMonitorLab) Cert() Cert             { return CertCKS }
func (l *CKSSysdigMonitorLab) DomainWeight() int      { return 20 }
func (l *CKSSysdigMonitorLab) Tags() []string {
	return []string{"cks", "sysdig", "monitoring", "runtime-security"}
}

func (l *CKSSysdigMonitorLab) Description() string {
	return `The cluster needs a comprehensive monitoring solution for security events.
Sysdig provides deep visibility into system calls and container behavior.

Your task: Install Sysdig agent in the 'sysdig-agent' namespace using Helm
to enable runtime security monitoring and threat detection.`
}

func (l *CKSSysdigMonitorLab) Hints() []string {
	return []string{
		"Add Sysdig Helm repository",
		"Create a secret with your Sysdig access key",
		"Install the agent with proper configuration",
	}
}

func (l *CKSSysdigMonitorLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSSysdigMonitorLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSSysdigMonitorLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "sysdig-agent", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get sysdig pods: %w", err)
	}
	if strings.Contains(output, "sysdig-agent") {
		return nil
	}
	return fmt.Errorf("sysdig agent not deployed")
}

func (l *CKSSysdigMonitorLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Add Sysdig Helm repo", Command: "helm repo add sysdig https://charts.sysdig.com/ && helm repo update"},
		{Description: "Create access key secret", Command: "kubectl create secret generic sysdig-agent --from-literal=access-key=YOUR_ACCESS_KEY -n sysdig-agent --create-namespace"},
		{Description: "Install Sysdig agent", Command: "helm install sysdig-agent sysdig/sysdig-agent -n sysdig-agent --set sysdig.accessKey=YOUR_ACCESS_KEY"},
		{Description: "Verify", Command: "kubectl get pods -n sysdig-agent"},
	}
}
