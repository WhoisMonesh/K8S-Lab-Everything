package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&EtcdBackupRestoreLab{})
}

type EtcdBackupRestoreLab struct {
	BaseLab
}

func (l *EtcdBackupRestoreLab) ID() string             { return "cka_etcd_backup_restore" }
func (l *EtcdBackupRestoreLab) Title() string          { return "Backup and Restore etcd Cluster" }
func (l *EtcdBackupRestoreLab) Category() Category     { return CategoryClusterArchitecture }
func (l *EtcdBackupRestoreLab) Difficulty() Difficulty { return DifficultyHard }
func (l *EtcdBackupRestoreLab) EstimatedTime() int     { return 30 }
func (l *EtcdBackupRestoreLab) Tags() []string {
	return []string{"etcd", "backup", "restore", "disaster-recovery"}
}
func (l *EtcdBackupRestoreLab) Cert() Cert        { return CertCKA }
func (l *EtcdBackupRestoreLab) DomainWeight() int { return 25 }

func (l *EtcdBackupRestoreLab) Description() string {
	return `Create a backup of the etcd cluster and verify it can be restored.
The backup should be saved to /tmp/etcd-backup.db and the restore should be
tested to ensure data integrity.`
}

func (l *EtcdBackupRestoreLab) Hints() []string {
	return []string{
		"Use etcdctl snapshot save to create the backup",
		"Ensure you have the correct CA and cert files",
		"Use etcdctl snapshot status to verify the backup",
	}
}

func (l *EtcdBackupRestoreLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *EtcdBackupRestoreLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *EtcdBackupRestoreLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "-n", "kube-system",
		"etcd-master", "--", "etcdctl", "snapshot", "status",
		"--endpoints=https://127.0.0.1:2379",
		"--cacert=/etc/kubernetes/pki/etcd/ca.crt",
		"--cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt",
		"--key=/etc/kubernetes/pki/etcd/healthcheck-client.key",
		"/tmp/etcd-backup.db")
	if err != nil {
		return fmt.Errorf("backup verification failed: %w", err)
	}
	if output == "" {
		return fmt.Errorf("backup file not found or empty")
	}
	return nil
}

func (l *EtcdBackupRestoreLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create etcd backup", Command: "ETCDCTL_API=3 etcdctl snapshot save /tmp/etcd-backup.db --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt --key=/etc/kubernetes/pki/etcd/healthcheck-client.key"},
		{Description: "Verify backup", Command: "ETCDCTL_API=3 etcdctl snapshot status /tmp/etcd-backup.db --write-table"},
		{Description: "Restore if needed", Command: "ETCDCTL_API=3 etcdctl snapshot restore /tmp/etcd-backup.db --data-dir=/var/lib/etcd-restored"},
	}
}
