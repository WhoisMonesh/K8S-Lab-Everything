package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADHelmInstallChartLab{})
}

type CKADHelmInstallChartLab struct {
	BaseLab
}

func (l *CKADHelmInstallChartLab) ID() string             { return "ckad_helm_install_chart" }
func (l *CKADHelmInstallChartLab) Title() string          { return "Install Helm Chart" }
func (l *CKADHelmInstallChartLab) Category() Category     { return CategoryAppDeployment }
func (l *CKADHelmInstallChartLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADHelmInstallChartLab) Cert() Cert             { return CertCKAD }
func (l *CKADHelmInstallChartLab) DomainWeight() int      { return 20 }
func (l *CKADHelmInstallChartLab) EstimatedTime() int     { return 20 }
func (l *CKADHelmInstallChartLab) Tags() []string {
	return []string{"helm", "charts", "deployment"}
}

func (l *CKADHelmInstallChartLab) Description() string {
	return `Deploy an application using Helm. The nginx-ingress controller chart
needs to be installed in the ingress-nginx namespace.

Your task: Add the ingress-nginx Helm repository and install the chart.`
}

func (l *CKADHelmInstallChartLab) Hints() []string {
	return []string{
		"Use helm repo add to add the repository",
		"Use helm install to deploy the chart",
		"Create the namespace if it doesn't exist",
	}
}

func (l *CKADHelmInstallChartLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADHelmInstallChartLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADHelmInstallChartLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "ingress-nginx",
		"-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get pods: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no pods found in ingress-nginx namespace")
	}
	return nil
}

func (l *CKADHelmInstallChartLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Add Helm repository", Command: "helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx"},
		{Description: "Update repos", Command: "helm repo update"},
		{Description: "Create namespace", Command: "kubectl create namespace ingress-nginx"},
		{Description: "Install chart", Command: "helm install ingress-nginx ingress-nginx/ingress-nginx -n ingress-nginx"},
	}
}
