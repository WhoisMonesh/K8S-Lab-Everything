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
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-apiserver", "-o", "jsonpath={.items[0].spec.containers[0].args}")
	if err != nil {
		return fmt.Errorf("failed to get apiserver args: %w", err)
	}
	if strings.Contains(output, "audit-policy-file") {
		return nil
	}
	return fmt.Errorf("audit policy not configured")
}

func (l *CKSAuditLogPolicyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create audit policy", Command: `sudo tee /etc/kubernetes/audit-policy.yaml <<EOF
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: RequestResponse
  resources:
  - group: ""
    resources: ["pods", "secrets"]
- level: Metadata
  resources:
- level: None
  nonResourceURLs: ["/healthz*", "/readyz*", "/livez*"]
EOF`},
		{Description: "Add to API server manifest", Command: "Edit /etc/kubernetes/manifests/kube-apiserver.yaml and add --audit-policy-file=/etc/kubernetes/audit-policy.yaml"},
		{Description: "Restart API server", Command: "sudo mv /etc/kubernetes/manifests/kube-apiserver.yaml /tmp/ && sleep 10 && sudo mv /tmp/kube-apiserver.yaml /etc/kubernetes/manifests/"},
	}
}
