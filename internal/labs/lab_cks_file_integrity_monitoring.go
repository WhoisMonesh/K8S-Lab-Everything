package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSFileIntegrityMonitoringLab{})
}

type CKSFileIntegrityMonitoringLab struct {
	BaseLab
}

func (l *CKSFileIntegrityMonitoringLab) ID() string             { return "cks_file_integrity_monitoring" }
func (l *CKSFileIntegrityMonitoringLab) Title() string          { return "Monitor File Integrity" }
func (l *CKSFileIntegrityMonitoringLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSFileIntegrityMonitoringLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSFileIntegrityMonitoringLab) EstimatedTime() int     { return 25 }
func (l *CKSFileIntegrityMonitoringLab) Cert() Cert             { return CertCKS }
func (l *CKSFileIntegrityMonitoringLab) DomainWeight() int      { return 20 }
func (l *CKSFileIntegrityMonitoringLab) Tags() []string {
	return []string{"cks", "file-integrity", "fim", "monitoring"}
}

func (l *CKSFileIntegrityMonitoringLab) Description() string {
	return `There is no file integrity monitoring on critical system files. Attackers
could modify binaries, configurations, or certificates without detection.

Your task: Deploy the File Integrity Operator (FIO) to monitor critical
files in /etc/kubernetes and /etc/ssl directories.`
}

func (l *CKSFileIntegrityMonitoringLab) Hints() []string {
	return []string{
		"Install the File Integrity Operator",
		"Create a FileIntegrity CRD",
		"Configure monitored paths for critical files",
	}
}

func (l *CKSFileIntegrityMonitoringLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSFileIntegrityMonitoringLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSFileIntegrityMonitoringLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "fileintegrity", "-A", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get file integrity: %w", err)
	}
	if strings.Contains(output, "fileintegrity") {
		return nil
	}
	return fmt.Errorf("file integrity monitoring not configured")
}

func (l *CKSFileIntegrityMonitoringLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Install File Integrity Operator", Command: "kubectl create -f https://raw.githubusercontent.com/openshift/file-integrity-operator/master/deploy/crds/fileintegrityalpha.crd.yaml"},
		{Description: "Create FileIntegrity CR", Command: `cat <<EOF | kubectl apply -f -
apiVersion: fileintegrityalpha.openshift.com/v1alpha1
kind: FileIntegrity
metadata:
  name: fileintegrity
  namespace: openshift-file-integrity
spec:
  config:
  - name: config
    data: |
      [Main]
      output=/var/log/file-integrity/aide.log
      [etc]
      dir=/etc
EOF`},
		{Description: "Verify", Command: "kubectl get fileintegrity -n openshift-file-integrity"},
	}
}
