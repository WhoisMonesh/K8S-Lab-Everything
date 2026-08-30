package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&APIServerAuditLogDisabled{})
}

type APIServerAuditLogDisabled struct {
	BaseLab
}

func (l *APIServerAuditLogDisabled) ID() string             { return "api_server_audit_log_disabled" }
func (l *APIServerAuditLogDisabled) Title() string          { return "API Server Audit Logging Disabled" }
func (l *APIServerAuditLogDisabled) Category() Category     { return CategoryControlPlane }
func (l *APIServerAuditLogDisabled) Difficulty() Difficulty { return DifficultyHard }
func (l *APIServerAuditLogDisabled) EstimatedTime() int     { return 25 }
func (l *APIServerAuditLogDisabled) Tags() []string {
	return []string{"audit", "api-server", "security"}
}

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
	// On a stock kind cluster audit logging is not enabled by default, which IS
	// the broken state this lab asks the learner to fix. Nothing to mutate here.
	return nil
}

func (l *APIServerAuditLogDisabled) Verify(ctx context.Context, kubeconfigPath string) error {
	node, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	// The audit policy file must exist on the control-plane node.
	policy, perr := dockerCommand(node, "cat /etc/kubernetes/audit-policy.yaml 2>/dev/null")
	if perr != nil || strings.TrimSpace(policy) == "" {
		return fmt.Errorf("audit policy file not found at /etc/kubernetes/audit-policy.yaml on %s", node)
	}
	if !strings.Contains(policy, "level:") {
		return fmt.Errorf("audit policy on %s has no rules", node)
	}
	// The API server manifest must enable audit logging via the policy file.
	manifest, merr := dockerCommand(node,
		"grep -c '--audit-policy-file' /etc/kubernetes/manifests/kube-apiserver.yaml")
	if merr != nil || strings.TrimSpace(manifest) == "0" {
		return fmt.Errorf("API server is not configured with --audit-policy-file")
	}
	return nil
}

func (l *APIServerAuditLogDisabled) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check the API server manifest for audit args", Command: "docker exec -it <cluster>-control-plane bash"},
		{Description: "Confirm audit logging is absent", Command: "grep -c audit /etc/kubernetes/manifests/kube-apiserver.yaml"},
		{Description: "Create the audit policy on the control-plane node", Command: "cat > /etc/kubernetes/audit-policy.yaml <<'EOF'\napiVersion: audit.k8s.io/v1\nkind: Policy\nrules:\n- level: Metadata\n  resources:\n  - group: \"\"\n    resources: [\"pods\", \"secrets\"]\nEOF"},
		{Description: "Enable audit logging in the API server manifest", Command: "sed -i '/--audit-log-path/a    - --audit-policy-file=/etc/kubernetes/audit-policy.yaml' /etc/kubernetes/manifests/kube-apiserver.yaml"},
		{Description: "Exit and let the static pod restart", Command: "exit && kubectl get pods -n kube-system -l component=kube-apiserver"},
	}
}
