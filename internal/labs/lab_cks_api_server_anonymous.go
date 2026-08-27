package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSAPIServerAnonymousLab{})
}

type CKSAPIServerAnonymousLab struct {
	BaseLab
}

func (l *CKSAPIServerAnonymousLab) ID() string             { return "cks_api_server_anonymous" }
func (l *CKSAPIServerAnonymousLab) Title() string          { return "Disable Anonymous Authentication" }
func (l *CKSAPIServerAnonymousLab) Category() Category     { return CategoryClusterSetupCKS }
func (l *CKSAPIServerAnonymousLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSAPIServerAnonymousLab) EstimatedTime() int     { return 20 }
func (l *CKSAPIServerAnonymousLab) Cert() Cert             { return CertCKS }
func (l *CKSAPIServerAnonymousLab) DomainWeight() int      { return 15 }
func (l *CKSAPIServerAnonymousLab) Tags() []string {
	return []string{"cks", "api-server", "anonymous", "authentication"}
}

func (l *CKSAPIServerAnonymousLab) Description() string {
	return `The API server allows anonymous requests, which means unauthenticated users
can query the cluster API. This is a security risk.

Your task: Disable anonymous authentication on the API server by setting
--anonymous-auth=false in the kube-apiserver manifest.`
}

func (l *CKSAPIServerAnonymousLab) Hints() []string {
	return []string{
		"Check the kube-apiserver manifest at /etc/kubernetes/manifests/",
		"Look for the --anonymous-auth flag",
		"Change its value to false",
	}
}

func (l *CKSAPIServerAnonymousLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSAPIServerAnonymousLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSAPIServerAnonymousLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-apiserver", "-o", "jsonpath={.items[0].spec.containers[0].args}")
	if err != nil {
		return fmt.Errorf("failed to get apiserver args: %w", err)
	}
	if strings.Contains(output, "anonymous-auth=false") {
		return nil
	}
	return fmt.Errorf("anonymous auth not disabled")
}

func (l *CKSAPIServerAnonymousLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current anonymous-auth setting", Command: "cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep anonymous-auth"},
		{Description: "Set anonymous-auth to false", Command: "sudo sed -i 's/--anonymous-auth=true/--anonymous-auth=false/' /etc/kubernetes/manifests/kube-apiserver.yaml"},
		{Description: "Verify API server restart", Command: "kubectl get pods -n kube-system -l component=kube-apiserver"},
	}
}
