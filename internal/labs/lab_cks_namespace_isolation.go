package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSNamespaceIsolationLab{})
}

type CKSNamespaceIsolationLab struct {
	BaseLab
}

func (l *CKSNamespaceIsolationLab) ID() string             { return "cks_namespace_isolation" }
func (l *CKSNamespaceIsolationLab) Title() string          { return "Implement Namespace Isolation" }
func (l *CKSNamespaceIsolationLab) Category() Category     { return CategoryClusterHardening }
func (l *CKSNamespaceIsolationLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSNamespaceIsolationLab) EstimatedTime() int     { return 20 }
func (l *CKSNamespaceIsolationLab) Cert() Cert             { return CertCKS }
func (l *CKSNamespaceIsolationLab) DomainWeight() int      { return 15 }
func (l *CKSNamespaceIsolationLab) Tags() []string {
	return []string{"cks", "namespace", "isolation", "resource-quota", "security"}
}

func (l *CKSNamespaceIsolationLab) Description() string {
	return `The 'prod' namespace has no resource limits or quotas. Pods can consume
unlimited resources, potentially affecting other namespaces.

Your task: Create a ResourceQuota in the 'prod' namespace that limits:
- Total CPU to 4 cores
- Total memory to 4Gi
- Maximum 10 pods
- Maximum 5 services`
}

func (l *CKSNamespaceIsolationLab) Hints() []string {
	return []string{
		"Use ResourceQuota to set namespace limits",
		"Define hard limits for cpu, memory, pods, and services",
		"Optionally create LimitRange for default pod limits",
	}
}

func (l *CKSNamespaceIsolationLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNamespaceIsolationLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: prod
`
	return kubectlApply(ctx, kubeconfigPath, ns)
}

func (l *CKSNamespaceIsolationLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "resourcequota", "-n", "prod", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get resource quota: %w", err)
	}
	if !strings.Contains(output, "hard") {
		return fmt.Errorf("resource quota not configured")
	}
	if !strings.Contains(output, "pods") || !strings.Contains(output, "services") {
		return fmt.Errorf("resource quota missing pod/service limits")
	}
	return nil
}

func (l *CKSNamespaceIsolationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create ResourceQuota", Command: `cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ResourceQuota
metadata:
  name: prod-quota
  namespace: prod
spec:
  hard:
    requests.cpu: "2"
    requests.memory: 2Gi
    limits.cpu: "4"
    limits.memory: 4Gi
    pods: "10"
    services: "5"
EOF`},
		{Description: "Verify quota", Command: "kubectl get resourcequota -n prod"},
	}
}
