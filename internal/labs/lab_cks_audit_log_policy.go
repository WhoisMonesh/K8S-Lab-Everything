package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSAuditLogPolicyLab{})
}

type CKSAuditLogPolicyLab struct {
	BaseLab
}

func (l *CKSAuditLogPolicyLab) ID() string             { return "cks_audit_log_policy" }
func (l *CKSAuditLogPolicyLab) Title() string          { return "Configure Audit Logging Policy" }
func (l *CKSAuditLogPolicyLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSAuditLogPolicyLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSAuditLogPolicyLab) EstimatedTime() int     { return 25 }
func (l *CKSAuditLogPolicyLab) Cert() Cert             { return CertCKS }
func (l *CKSAuditLogPolicyLab) DomainWeight() int      { return 20 }
func (l *CKSAuditLogPolicyLab) Tags() []string {
	return []string{"cks", "audit-logging", "monitoring", "security"}
}

func (l *CKSAuditLogPolicyLab) Description() string {
	return `The cluster does not have an audit logging policy configured. Without audit
logging, there is no record of API server activity for security investigations.

Your task: Create an audit policy that:
1. Logs RequestResponse level for pods and secrets
2. Logs Metadata level for all other requests
3. Logs None level for health check endpoints
Configure the API server to use this policy.`
}

func (l *CKSAuditLogPolicyLab) Hints() []string {
	return []string{
		"Create /etc/kubernetes/audit-policy.yaml",
		"Define rules with different log levels",
		"Add --audit-policy-file flag to kube-apiserver",
	}
}

func (l *CKSAuditLogPolicyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSAuditLogPolicyLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSAuditLogPolicyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	node, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	// The API server manifest must enable audit logging via the policy file.
	manifest, merr := dockerCommand(node,
		"grep -c '--audit-policy-file' /etc/kubernetes/manifests/kube-apiserver.yaml")
	if merr != nil || strings.TrimSpace(manifest) == "0" {
		return fmt.Errorf("API server is not configured with --audit-policy-file")
	}
	// And the policy file must carry the required log levels.
	policy, perr := dockerCommand(node, "cat /etc/kubernetes/audit-policy.yaml 2>/dev/null")
	if perr != nil || strings.TrimSpace(policy) == "" {
		return fmt.Errorf("audit policy file not found at /etc/kubernetes/audit-policy.yaml on %s", node)
	}
	if !strings.Contains(policy, "RequestResponse") {
		return fmt.Errorf("audit policy does not log RequestResponse level for pods/secrets")
	}
	return nil
}

func (l *CKSAuditLogPolicyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Enter the control-plane node shell (kind has no SSH)", Command: "docker exec -it <cluster>-control-plane bash"},
		{Description: "Create the audit policy", Command: "cat > /etc/kubernetes/audit-policy.yaml <<'EOF'\napiVersion: audit.k8s.io/v1\nkind: Policy\nrules:\n- level: RequestResponse\n  resources:\n  - group: \"\"\n    resources: [\"pods\", \"secrets\"]\n- level: Metadata\n  resources:\n  - group: \"\"\n    resources: []\n- level: None\n  nonResourceURLs: [\"/healthz*\", \"/readyz*\", \"/livez*\"]\nEOF"},
		{Description: "Add the flag to the API server manifest", Command: "sed -i '/--audit-log-path/a    - --audit-policy-file=/etc/kubernetes/audit-policy.yaml' /etc/kubernetes/manifests/kube-apiserver.yaml"},
		{Description: "Let the static pod restart and verify", Command: "exit && kubectl get pods -n kube-system -l component=kube-apiserver"},
	}
}
