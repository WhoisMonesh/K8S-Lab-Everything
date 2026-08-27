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
	output, err := kubectl(ctx, kubeconfigPath, "exec", "-n", "kube-system",
		" kube-apiserver-master", "--", "cat", "/etc/kubernetes/audit-policy.yaml")
	if err != nil {
		return fmt.Errorf("audit policy not found: %w", err)
	}
	if !strings.Contains(output, "RequestResponse") {
		return fmt.Errorf("audit policy missing RequestResponse level")
	}
	return nil
}

func (l *APIServerAuditPolicyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create audit policy", Command: "cat <<EOF | sudo tee /etc/kubernetes/audit-policy.yaml\napiVersion: audit.k8s.io/v1\nkind: Policy\nrules:\n- level: RequestResponse\n  resources:\n  - group: \"\"\n    resources: [\"pods\", \"secrets\"]\n- level: Metadata\n  resources:\n- level: None\n  nonResourceURLs: [\"/healthz*\"]\nEOF"},
		{Description: "Add audit policy to API server", Command: "sudo sed -i '/--audit-policy-file/a\\\\    - --audit-policy-file=/etc/kubernetes/audit-policy.yaml' /etc/kubernetes/manifests/kube-apiserver.yaml"},
		{Description: "Verify API server restarted", Command: "kubectl get pods -n kube-system -l component=kube-apiserver"},
	}
}
