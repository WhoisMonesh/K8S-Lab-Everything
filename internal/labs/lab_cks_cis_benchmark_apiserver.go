package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&CKSCISBenchmarkAPIServerLab{})
}

type CKSCISBenchmarkAPIServerLab struct {
	BaseLab
}

func (l *CKSCISBenchmarkAPIServerLab) ID() string             { return "cks_cis_benchmark_apiserver" }
func (l *CKSCISBenchmarkAPIServerLab) Title() string          { return "Harden API Server with CIS Benchmarks" }
func (l *CKSCISBenchmarkAPIServerLab) Category() Category     { return CategoryClusterSetupCKS }
func (l *CKSCISBenchmarkAPIServerLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSCISBenchmarkAPIServerLab) EstimatedTime() int     { return 30 }
func (l *CKSCISBenchmarkAPIServerLab) Cert() Cert             { return CertCKS }
func (l *CKSCISBenchmarkAPIServerLab) DomainWeight() int      { return 15 }
func (l *CKSCISBenchmarkAPIServerLab) Tags() []string {
	return []string{"cks", "api-server", "cis-benchmark", "hardening"}
}

func (l *CKSCISBenchmarkAPIServerLab) Description() string {
	return `The API server configuration lacks several CIS benchmark security settings.
Configure the API server to enforce:
- --anonymous-auth=false
- --authorization-mode=Node,RBAC
- --enable-admission-plugins=NodeRestriction`
}

func (l *CKSCISBenchmarkAPIServerLab) Hints() []string {
	return []string{
		"Check the kube-apiserver manifest in /etc/kubernetes/manifests/",
		"CIS benchmarks require specific flags on the API server",
		"Pods restart automatically when the manifest changes",
	}
}

func (l *CKSCISBenchmarkAPIServerLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSCISBenchmarkAPIServerLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSCISBenchmarkAPIServerLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-apiserver", "-o", "jsonpath={.items[0].spec.containers[0].args}")
	if err != nil {
		return fmt.Errorf("failed to get apiserver args: %w", err)
	}
	if !containsAny(output, "anonymous-auth=false") {
		return fmt.Errorf("anonymous-auth not disabled")
	}
	if !containsAny(output, "authorization-mode=Node,RBAC") {
		return fmt.Errorf("authorization-mode not set correctly")
	}
	return nil
}

func (l *CKSCISBenchmarkAPIServerLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current API server config", Command: "cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep -E 'anonymous-auth|authorization-mode'"},
		{Description: "Add required flags", Command: "sudo sed -i '/--anonymous-auth/a\\    - --anonymous-auth=false\\n    - --authorization-mode=Node,RBAC\\n    - --enable-admission-plugins=NodeRestriction' /etc/kubernetes/manifests/kube-apiserver.yaml"},
		{Description: "Verify API server restart", Command: "kubectl get pods -n kube-system -l component=kube-apiserver"},
	}
}
