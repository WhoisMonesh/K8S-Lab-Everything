package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&EtcdRecoveryLab{})
}

type EtcdRecoveryLab struct {
	BaseLab
}

func (l *EtcdRecoveryLab) ID() string {
	return "etcd_recovery"
}

func (l *EtcdRecoveryLab) Title() string {
	return "Advanced etcd Recovery"
}

func (l *EtcdRecoveryLab) Category() Category {
	return CategoryClusterArchitecture
}

func (l *EtcdRecoveryLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *EtcdRecoveryLab) Description() string {
	return `The etcd database has a disk space alarm triggered due to excessive
fragmentation. The cluster is experiencing issues with resource
creation and updates.

Your task: Defragment and compact the etcd database to clear the alarm
and restore normal cluster operations.`
}

func (l *EtcdRecoveryLab) Hints() []string {
	return []string{
		"Check etcd status with etcdctl endpoint status",
		"Use etcdctl defrag to compact the database",
		"Check for alarms with etcdctl alarm list",
	}
}

func (l *EtcdRecoveryLab) EstimatedTime() int {
	return 30
}

func (l *EtcdRecoveryLab) Tags() []string {
	return []string{"etcd", "backup", "recovery", "cluster-architecture", "control-plane"}
}

func (l *EtcdRecoveryLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *EtcdRecoveryLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("getting control plane node: %w", err)
	}

	etcdctl := func(args ...string) {
		baseArgs := []string{"etcdctl",
			"--endpoints=https://127.0.0.1:2379",
			"--cacert=/etc/kubernetes/pki/etcd/ca.crt",
			"--cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt",
			"--key=/etc/kubernetes/pki/etcd/healthcheck-client.key"}
		allArgs := append(baseArgs, args...)
		dockerExec(ctx, nodeName, allArgs...)
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < 100; j++ {
			kubectl(ctx, kubeconfigPath, "create", "configmap",
				fmt.Sprintf("filler-%d-%d", i, j),
				"--from-literal=key=large-tenant-value",
				"-n", "default")
		}
	}

	etcdctl("alarm", "add", "NOSPACE")

	return nil
}

func (l *EtcdRecoveryLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(5 * time.Second)

	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return nil
	}

	output, _ := dockerExec(ctx, nodeName, "etcdctl",
		"--endpoints=https://127.0.0.1:2379",
		"--cacert=/etc/kubernetes/pki/etcd/ca.crt",
		"--cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt",
		"--key=/etc/kubernetes/pki/etcd/healthcheck-client.key",
		"alarm", "list")

	if strings.Contains(output, "NOSPACE") {
		return nil
	}

	return fmt.Errorf("NOSPACE alarm not found")
}

func (l *EtcdRecoveryLab) Verify(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)

	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("getting control plane node: %w", err)
	}

	output, _ := dockerExec(ctx, nodeName, "etcdctl",
		"--endpoints=https://127.0.0.1:2379",
		"--cacert=/etc/kubernetes/pki/etcd/ca.crt",
		"--cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt",
		"--key=/etc/kubernetes/pki/etcd/healthcheck-client.key",
		"alarm", "list")

	if strings.Contains(output, "NOSPACE") {
		return fmt.Errorf("NOSPACE alarm still present")
	}

	return nil
}

func (l *EtcdRecoveryLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check etcd alarms",
			Command:     "docker exec <control-plane> etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt --key=/etc/kubernetes/pki/etcd/healthcheck-client.key alarm list",
			Notes:       "Shows NOSPACE alarm",
		},
		{
			Description: "Defragment etcd",
			Command:     "docker exec <control-plane> etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt --key=/etc/kubernetes/pki/etcd/healthcheck-client.key defrag",
			Notes:       "Compacts and defragments the database",
		},
		{
			Description: "Clear alarms",
			Command:     "docker exec <control-plane> etcdctl --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt --key=/etc/kubernetes/pki/etcd/healthcheck-client.key alarm disarm NOSPACE",
			Notes:       "Disarms the NOSPACE alarm",
		},
		{
			Description: "Verify etcd is healthy",
			Command:     "kubectl get pods -n kube-system | grep etcd",
			Notes:       "etcd should be running and healthy",
		},
	}
}
