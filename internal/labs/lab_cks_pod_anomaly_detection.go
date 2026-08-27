package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSPodAnomalyDetectionLab{})
}

type CKSPodAnomalyDetectionLab struct {
	BaseLab
}

func (l *CKSPodAnomalyDetectionLab) ID() string             { return "cks_pod_anomaly_detection" }
func (l *CKSPodAnomalyDetectionLab) Title() string          { return "Detect Pod Anomalies" }
func (l *CKSPodAnomalyDetectionLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSPodAnomalyDetectionLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSPodAnomalyDetectionLab) EstimatedTime() int     { return 25 }
func (l *CKSPodAnomalyDetectionLab) Cert() Cert             { return CertCKS }
func (l *CKSPodAnomalyDetectionLab) DomainWeight() int      { return 20 }
func (l *CKSPodAnomalyDetectionLab) Tags() []string {
	return []string{"cks", "anomaly-detection", "pod", "monitoring"}
}

func (l *CKSPodAnomalyDetectionLab) Description() string {
	return `Pods that suddenly change behavior, restart frequently, or consume
abnormal resources may indicate a security compromise. There is currently
no automated anomaly detection.

Your task: Create a Kubernetes CronJob that checks for pods with high
restart counts (>5) and reports them as potential anomalies.`
}

func (l *CKSPodAnomalyDetectionLab) Hints() []string {
	return []string{
		"Use kubectl get pods to check restart counts",
		"Parse output to find pods with restarts > 5",
		"Log anomalies to a ConfigMap",
	}
}

func (l *CKSPodAnomalyDetectionLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSPodAnomalyDetectionLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSPodAnomalyDetectionLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "cronjob", "-A", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get cronjobs: %w", err)
	}
	if strings.Contains(output, "pod-anomaly") {
		return nil
	}
	return fmt.Errorf("pod anomaly detection CronJob not found")
}

func (l *CKSPodAnomalyDetectionLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create anomaly detection CronJob", Command: `cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: CronJob
metadata:
  name: pod-anomaly
  namespace: monitoring
spec:
  schedule: "*/5 * * * *"
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
              echo "=== Pod Anomaly Detection ==="
              kubectl get pods --all-namespaces -o json | \\
              jq -r '.items[] | select(.status.containerStatuses[].restartCount > 5) | .metadata.namespace + "/" + .metadata.name'
          restartPolicy: OnFailure
EOF`},
		{Description: "Verify", Command: "kubectl get cronjob pod-anomaly -n monitoring"},
	}
}
