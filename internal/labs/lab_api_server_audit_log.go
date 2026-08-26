package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&APIServerAuditLogDisabled{})
}

type APIServerAuditLogDisabled struct {
	BaseLab
}

func (l *APIServerAuditLogDisabled) ID() string            { return "api_server_audit_log_disabled" }
func (l *APIServerAuditLogDisabled) Title() string         { return "API Server Audit Logging Disabled" }
func (l *APIServerAuditLogDisabled) Category() Category    { return CategoryControlPlane }
func (l *APIServerAuditLogDisabled) Difficulty() Difficulty { return DifficultyHard }
func (l *APIServerAuditLogDisabled) EstimatedTime() int    { return 25 }
func (l *APIServerAuditLogDisabled) Tags() []string        { return []string{"audit", "api-server", "security"} }

func (l *APIServerAuditLogDisabled) Description() string {
	return `Audit logging is disabled on the API server. Enable audit logging with a policy that logs all requests at the Metadata level.
The audit policy file should be at /etc/kubernetes/audit-policy.yaml.`
}

func (l *APIServerAuditLogDisabled) Hints() []string {
	return []string{
		"Check the API server static pod manifest",
		"Look for --audit-policy-file flag",
		"Create an audit policy file if missing",
	}
}

func (l *APIServerAuditLogDisabled) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *APIServerAuditLogDisabled) Break(ctx context.Context, kubeconfigPath string) error {
	policy := `apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: None
  resources:
  - group: ""
    resources: ["events"]`
	return kubectlApply(ctx, kubeconfigPath, policy)
}

func (l *APIServerAuditLogDisabled) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system", "-l", "component=kube-apiserver",
		"-o", "jsonpath={.items[0].spec.containers[0].args}")
	if err != nil {
		return err
	}
	if !containsAny(output, "audit-policy-file") {
		return fmt.Errorf("audit-policy-file flag not found")
	}
	return nil
}

func (l *APIServerAuditLogDisabled) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check API server args", Command: "kubectl describe pods -n kube-system -l component=kube-apiserver"},
		{Description: "Create audit policy", Command: "sudo tee /etc/kubernetes/audit-policy.yaml <<EOF\napiVersion: audit.k8s.io/v1\nkind: Policy\nrules:\n- level: Metadata\nEOF"},
		{Description: "Add audit flag to API server", Command: "Edit /etc/kubernetes/manifests/kube-apiserver.yaml and add --audit-policy-file=/etc/kubernetes/audit-policy.yaml"},
	}
}
