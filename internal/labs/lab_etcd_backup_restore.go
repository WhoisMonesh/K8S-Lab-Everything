package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&EtcdBackupRestoreLab{})
}

type EtcdBackupRestoreLab struct {
	BaseLab
}

func (l *EtcdBackupRestoreLab) ID() string {
	return "etcd_backup_restore"
}

func (l *EtcdBackupRestoreLab) Title() string {
	return "Etcd Backup and Restore"
}

func (l *EtcdBackupRestoreLab) Category() Category {
	return CategoryControlPlane
}

func (l *EtcdBackupRestoreLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *EtcdBackupRestoreLab) Description() string {
	return `A critical ConfigMap has been accidentally deleted from the cluster.
Fortunately, there should be an etcd backup available. However, the backup
location is not properly documented.

Your task: Find the etcd backup, verify it exists, and prepare the commands
needed to restore the cluster state. Create a marker ConfigMap to indicate
you've successfully prepared the restore procedure.`
}

func (l *EtcdBackupRestoreLab) Hints() []string {
	return []string{
		"Check the /var/lib/etcd-backup directory on the control plane node",
		"Use etcdctl to interact with etcd backups",
		"You'll need the etcd certificates for authentication",
		"The certificates are typically in /etc/kubernetes/pki/etcd/",
	}
}

func (l *EtcdBackupRestoreLab) EstimatedTime() int {
	return 30
}

func (l *EtcdBackupRestoreLab) Tags() []string {
	return []string{"etcd", "backup", "restore", "disaster-recovery", "control-plane"}
}

func (l *EtcdBackupRestoreLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	// Create a test ConfigMap that will be "deleted"
	testCM := `apiVersion: v1
kind: ConfigMap
metadata:
  name: critical-app-config
  namespace: default
data:
  app-mode: "production"
  database-host: "db.example.com"
  feature-flags: "analytics:on,reporting:on"
`
	return kubectlApply(ctx, kubeconfigPath, testCM)
}

func (l *EtcdBackupRestoreLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	containerName := nodeName

	// Create a backup directory and a simulated backup file
	_, err = dockerExec(ctx, containerName, "mkdir", "-p", "/var/lib/etcd-backup")
	if err != nil {
		return fmt.Errorf("creating backup directory: %w", err)
	}

	// Create a marker file to simulate a backup
	backupInfo := `Backup Information:
Location: /var/lib/etcd-backup/snapshot.db
Timestamp: 2024-01-15 10:30:00
Status: completed
Size: 2.3 MB
`
	cmd := fmt.Sprintf("cat > /var/lib/etcd-backup/backup-info.txt << 'EOF'\n%s\nEOF", backupInfo)
	_, err = dockerExec(ctx, containerName, "sh", "-c", cmd)
	if err != nil {
		return fmt.Errorf("creating backup info: %w", err)
	}

	// Create an actual etcdctl snapshot (simplified simulation)
	_, err = dockerExec(ctx, containerName, "sh", "-c",
		"echo 'etcd-snapshot-placeholder' > /var/lib/etcd-backup/snapshot.db")
	if err != nil {
		return fmt.Errorf("creating snapshot file: %w", err)
	}

	// Delete the critical ConfigMap
	_, err = kubectl(ctx, kubeconfigPath, "delete", "configmap",
		"critical-app-config", "-n", "default", "--ignore-not-found")
	if err != nil {
		return fmt.Errorf("deleting ConfigMap: %w", err)
	}

	return nil
}

func (l *EtcdBackupRestoreLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Verify the ConfigMap is deleted
	_, err := kubectl(ctx, kubeconfigPath, "get", "configmap",
		"critical-app-config", "-n", "default")
	if err == nil {
		return fmt.Errorf("ConfigMap should be deleted")
	}
	return nil
}

func (l *EtcdBackupRestoreLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if a restore-ready marker ConfigMap has been created
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap",
		"etcd-restore-ready", "-n", "kube-system",
		"-o", "jsonpath={.data.backup-verified}")
	if err != nil {
		return fmt.Errorf("restore-ready marker ConfigMap not found: %w", err)
	}

	if !strings.Contains(output, "true") {
		return fmt.Errorf("backup not verified in restore-ready ConfigMap")
	}

	// Optionally check if the backup location is documented
	location, err := kubectl(ctx, kubeconfigPath, "get", "configmap",
		"etcd-restore-ready", "-n", "kube-system",
		"-o", "jsonpath={.data.backup-location}")
	if err == nil && strings.Contains(location, "/var/lib/etcd-backup") {
		// Good - they documented the location
		return nil
	}

	return nil
}

func (l *EtcdBackupRestoreLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Verify the critical ConfigMap is missing",
			Command:     "kubectl get configmap critical-app-config -n default",
			Notes:       "This should return a 'not found' error",
		},
		{
			Description: "Access the control plane node",
			Command:     "docker exec -it <cluster-name>-control-plane bash",
			Notes:       "You need to access the node to check for backups",
		},
		{
			Description: "Look for etcd backups",
			Command:     "ls -la /var/lib/etcd-backup/",
			Notes:       "Check what backup files are available",
		},
		{
			Description: "Review the backup information",
			Command:     "cat /var/lib/etcd-backup/backup-info.txt",
			Notes:       "Get details about the backup",
		},
		{
			Description: "Verify the snapshot file exists",
			Command:     "ls -lh /var/lib/etcd-backup/snapshot.db",
			Notes:       "Confirm the backup file is present",
		},
		{
			Description: "Document the restore preparation (exit the container first)",
			Command: `kubectl create configmap etcd-restore-ready -n kube-system \
  --from-literal=backup-verified=true \
  --from-literal=backup-location=/var/lib/etcd-backup/snapshot.db`,
			Notes: "Create a marker ConfigMap indicating you've verified the backup and are ready to restore",
		},
		{
			Description: "Verify the solution",
			Command:     "kubectl get configmap etcd-restore-ready -n kube-system -o yaml",
			Notes:       "Confirm the restore-ready marker has been created correctly",
		},
	}
}
