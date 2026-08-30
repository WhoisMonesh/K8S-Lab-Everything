package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&APIServerAuditPolicyLab{})
}

type APIServerAuditPolicyLab struct {
	BaseLab
}

func (l *APIServerAuditPolicyLab) ID() string             { return "cka_api_server_audit_policy" }
func (l *APIServerAuditPolicyLab) Title() string          { return "Configure API Server Audit Logging" }
func (l *APIServerAuditPolicyLab) Category() Category     { return CategoryClusterArchitecture }
func (l *APIServerAuditPolicyLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *APIServerAuditPolicyLab) EstimatedTime() int     { return 25 }
func (l *APIServerAuditPolicyLab) Tags() []string {
	return []string{"api-server", "audit", "logging", "security"}
}
func (l *APIServerAuditPolicyLab) Cert() Cert        { return CertCKA }
func (l *APIServerAuditPolicyLab) DomainWeight() int { return 25 }

func (l *APIServerAuditPolicyLab) Description() string {
	return `The API server audit logging is not configured. Create an audit policy
that logs RequestResponse level for pods and secrets, and Metadata level for
all other requests. Configure the API server to use this policy.`
}

func (l *APIServerAuditPolicyLab) Hints() []string {
	return []string{
		"Create an audit policy YAML file",
		"Add --audit-policy-file flag to API server manifest",
		"Ensure the policy file is accessible to the API server",
	}
}

func (l *APIServerAuditPolicyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *APIServerAuditPolicyLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *APIServerAuditPolicyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	node, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	// The audit policy file lives on the control-plane node. Check it via
	// docker exec (kind nodes have no SSH) and confirm the API server is
	// actually configured with it.
	output, derr := dockerCommand(node, "cat /etc/kubernetes/audit-policy.yaml 2>/dev/null")
	if derr != nil || !strings.Contains(output, "RequestResponse") {
		return fmt.Errorf("audit policy file on %s is missing or lacks RequestResponse", node)
	}
	// API server manifest must reference the policy file.
	manifest, merr := dockerCommand(node, "grep -l 'audit-policy-file' /etc/kubernetes/manifests/kube-apiserver.yaml 2>/dev/null")
	if merr != nil || strings.TrimSpace(manifest) == "" {
		return fmt.Errorf("API server manifest does not enable --audit-policy-file")
	}
	return nil
}

func (l *APIServerAuditPolicyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Enter the control-plane node shell (kind has no SSH)", Command: "docker exec -it <cluster>-control-plane bash"},
		{Description: "Create the audit policy on the node", Command: "cat > /etc/kubernetes/audit-policy.yaml <<'EOF'\napiVersion: audit.k8s.io/v1\nkind: Policy\nrules:\n- level: RequestResponse\n  resources:\n  - group: \"\"\n    resources: [\"pods\", \"secrets\"]\n- level: Metadata\n  resources: []\n- level: None\n  nonResourceURLs: [\"/healthz*\"]\nEOF"},
		{Description: "Point the API server at the policy", Command: "sed -i '/--audit-log-path/a    - --audit-policy-file=/etc/kubernetes/audit-policy.yaml' /etc/kubernetes/manifests/kube-apiserver.yaml"},
		{Description: "Let the static pod restart, then verify", Command: "exit && kubectl get pods -n kube-system -l component=kube-apiserver"},
	}
}
