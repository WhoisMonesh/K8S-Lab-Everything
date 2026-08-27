package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSPrometheusAlertLab{})
}

type CKSPrometheusAlertLab struct {
	BaseLab
}

func (l *CKSPrometheusAlertLab) ID() string             { return "cks_prometheus_alert" }
func (l *CKSPrometheusAlertLab) Title() string          { return "Configure Prometheus Security Alerts" }
func (l *CKSPrometheusAlertLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSPrometheusAlertLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSPrometheusAlertLab) EstimatedTime() int     { return 25 }
func (l *CKSPrometheusAlertLab) Cert() Cert             { return CertCKS }
func (l *CKSPrometheusAlertLab) DomainWeight() int      { return 20 }
func (l *CKSPrometheusAlertLab) Tags() []string {
	return []string{"cks", "prometheus", "alerting", "monitoring"}
}

func (l *CKSPrometheusAlertLab) Description() string {
	return `The cluster monitoring does not have security-focused alerts. Critical events
like pod privilege escalation, unauthorized API access, and container escapes
are not detected.

Your task: Create a PrometheusRule resource with security alerts for:
1. Pods running as root
2. Failed authentication attempts
3. Unauthorized API calls`
}

func (l *CKSPrometheusAlertLab) Hints() []string {
	return []string{
		"Create a PrometheusRule CRD",
		"Define alert rules with appropriate thresholds",
		"Use kube-apiserver request metrics for auth failures",
	}
}

func (l *CKSPrometheusAlertLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSPrometheusAlertLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSPrometheusAlertLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "prometheusrule", "-A", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get prometheusrules: %w", err)
	}
	if strings.Contains(output, "security-alerts") {
		return nil
	}
	return fmt.Errorf("security alerts PrometheusRule not found")
}

func (l *CKSPrometheusAlertLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create security alerts rule", Command: `cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: security-alerts
  namespace: monitoring
spec:
  groups:
  - name: security
    rules:
    - alert: PodRunningAsRoot
      expr: kube_pod_container_status_running{namespace!="kube-system"} > 0 and on(pod) kube_pod_info{pod=~".*root.*"}
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "Pod running as root detected"
    - alert: HighAuthFailures
      expr: sum(rate(apiserver_audit_event_total{verb="authentication",code="401"}[5m])) > 0.1
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "High rate of authentication failures"
EOF`},
		{Description: "Verify", Command: "kubectl get prometheusrule -n monitoring"},
	}
}
