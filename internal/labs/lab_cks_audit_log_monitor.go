package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSAuditLogMonitorLab{})
}

type CKSAuditLogMonitorLab struct {
	BaseLab
}

func (l *CKSAuditLogMonitorLab) ID() string             { return "cks_audit_log_monitor" }
func (l *CKSAuditLogMonitorLab) Title() string          { return "Monitor Audit Log for Threats" }
func (l *CKSAuditLogMonitorLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSAuditLogMonitorLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSAuditLogMonitorLab) EstimatedTime() int     { return 25 }
func (l *CKSAuditLogMonitorLab) Cert() Cert             { return CertCKS }
func (l *CKSAuditLogMonitorLab) DomainWeight() int      { return 20 }
func (l *CKSAuditLogMonitorLab) Tags() []string {
	return []string{"cks", "audit-log", "monitoring", "threat-detection"}
}

func (l *CKSAuditLogMonitorLab) Description() string {
	return `Audit logs are being collected but there is no monitoring for suspicious
activities. Critical security events like unauthorized access attempts,
privilege escalation, and secret access go undetected.

Your task: Create a Kubernetes CronJob that periodically checks the audit log
for failed authentication attempts and generates alerts.`
}

func (l *CKSAuditLogMonitorLab) Hints() []string {
	return []string{
		"Create a CronJob that greps audit logs",
		"Look for 'rejected' or 'Unauthorized' entries",
		"Output alerts to a ConfigMap or stdout",
	}
}

func (l *CKSAuditLogMonitorLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSAuditLogMonitorLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSAuditLogMonitorLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "cronjob", "-A", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get cronjobs: %w", err)
	}
	if strings.Contains(output, "audit-monitor") {
		return nil
	}
	return fmt.Errorf("audit monitoring CronJob not found")
}

func (l *CKSAuditLogMonitorLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create audit monitoring CronJob", Command: `cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: CronJob
metadata:
  name: audit-monitor
  namespace: kube-system
spec:
  schedule: "*/5 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: audit-reader
          containers:
          - name: monitor
            image: busybox:1.36
            command:
            - sh
            - -c
            - |
              echo "=== Audit Log Security Check ==="
              echo "Checking for failed authentications..."
              cat /var/log/kubernetes/audit/audit.log | grep -c "Unauthorized" || echo "0"
          volumeMounts:
          - name: audit-log
            mountPath: /var/log/kubernetes/audit
            readOnly: true
          volumes:
          - name: audit-log
            hostPath:
              path: /var/log/kubernetes/audit
          restartPolicy: OnFailure
EOF`},
		{Description: "Verify", Command: "kubectl get cronjob audit-monitor -n kube-system"},
	}
}
