package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSSuspiciousActivityDetectionLab{})
}

type CKSSuspiciousActivityDetectionLab struct {
	BaseLab
}

func (l *CKSSuspiciousActivityDetectionLab) ID() string { return "cks_suspicious_activity_detect" }
func (l *CKSSuspiciousActivityDetectionLab) Title() string {
	return "Detect Suspicious Activity Patterns"
}
func (l *CKSSuspiciousActivityDetectionLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSSuspiciousActivityDetectionLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSSuspiciousActivityDetectionLab) EstimatedTime() int     { return 30 }
func (l *CKSSuspiciousActivityDetectionLab) Cert() Cert             { return CertCKS }
func (l *CKSSuspiciousActivityDetectionLab) DomainWeight() int      { return 20 }
func (l *CKSSuspiciousActivityDetectionLab) Tags() []string {
	return []string{"cks", "threat-detection", "suspicious-activity", "monitoring"}
}

func (l *CKSSuspiciousActivityDetectionLab) Description() string {
	return `There is no detection mechanism for suspicious patterns such as:
- Multiple failed authentication attempts
- Unusual API access patterns
- Unexpected container exec activity

Your task: Create a CronJob that monitors the Kubernetes audit log for
suspicious patterns and sends alerts to a ConfigMap.`
}

func (l *CKSSuspiciousActivityDetectionLab) Hints() []string {
	return []string{
		"Parse audit logs for failed auth events",
		"Look for exec operations in unusual namespaces",
		"Store alerts in a ConfigMap for review",
	}
}

func (l *CKSSuspiciousActivityDetectionLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSSuspiciousActivityDetectionLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSSuspiciousActivityDetectionLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "cronjob", "-A", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get cronjobs: %w", err)
	}
	if strings.Contains(output, "suspicious-activity") {
		return nil
	}
	return fmt.Errorf("suspicious activity monitoring CronJob not found")
}

func (l *CKSSuspiciousActivityDetectionLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create monitoring CronJob", Command: `cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: CronJob
metadata:
  name: suspicious-activity
  namespace: kube-system
spec:
  schedule: "*/10 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: monitor
            image: busybox:1.36
            command:
            - sh
            - -c
            - |
              echo "Checking for suspicious activity..."
              TIMESTAMP=$(date -u +%Y-%m-%dT%H:%M:%SZ)
              echo "\$TIMESTAMP: Security scan completed" >> /dev/stdout
          restartPolicy: OnFailure
EOF`},
		{Description: "Verify", Command: "kubectl get cronjob suspicious-activity -n kube-system"},
	}
}
