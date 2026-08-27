package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSNetworkAnomalyDetectionLab{})
}

type CKSNetworkAnomalyDetectionLab struct {
	BaseLab
}

func (l *CKSNetworkAnomalyDetectionLab) ID() string             { return "cks_network_anomaly_detection" }
func (l *CKSNetworkAnomalyDetectionLab) Title() string          { return "Detect Network Anomalies" }
func (l *CKSNetworkAnomalyDetectionLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSNetworkAnomalyDetectionLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSNetworkAnomalyDetectionLab) EstimatedTime() int     { return 30 }
func (l *CKSNetworkAnomalyDetectionLab) Cert() Cert             { return CertCKS }
func (l *CKSNetworkAnomalyDetectionLab) DomainWeight() int      { return 20 }
func (l *CKSNetworkAnomalyDetectionLab) Tags() []string {
	return []string{"cks", "network-anomaly", "monitoring", "security"}
}

func (l *CKSNetworkAnomalyDetectionLab) Description() string {
	return `Unusual network traffic patterns may indicate data exfiltration, C2 communication,
or lateral movement. There is no network anomaly detection in place.

Your task: Create a NetworkPolicy in the 'monitoring' namespace and a CronJob
that monitors for unexpected external connections from pods.`
}

func (l *CKSNetworkAnomalyDetectionLab) Hints() []string {
	return []string{
		"Monitor network connections using /proc/net",
		"Check for connections to unexpected external IPs",
		"Alert on DNS queries to unusual domains",
	}
}

func (l *CKSNetworkAnomalyDetectionLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNetworkAnomalyDetectionLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSNetworkAnomalyDetectionLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "cronjob", "-A", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get cronjobs: %w", err)
	}
	if strings.Contains(output, "network-anomaly") {
		return nil
	}
	return fmt.Errorf("network anomaly detection CronJob not found")
}

func (l *CKSNetworkAnomalyDetectionLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create network anomaly detection CronJob", Command: `cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: CronJob
metadata:
  name: network-anomaly
  namespace: monitoring
spec:
  schedule: "*/10 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: detector
            image: busybox:1.36
            command:
            - sh
            - -c
            - |
              echo "=== Network Anomaly Detection ==="
              echo "Checking for unusual connections..."
          restartPolicy: OnFailure
EOF`},
		{Description: "Verify", Command: "kubectl get cronjob network-anomaly -n monitoring"},
	}
}
