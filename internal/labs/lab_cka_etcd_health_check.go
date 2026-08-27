package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&EtcdHealthCheckLab{})
}

type EtcdHealthCheckLab struct {
	BaseLab
}

func (l *EtcdHealthCheckLab) ID() string { return "cka_etcd_health_check" }
func (l *EtcdHealthCheckLab) Title() string {
	return "Debug etcd Health Issues"
}
func (l *EtcdHealthCheckLab) Category() Category     { return CategoryTroubleshooting }
func (l *EtcdHealthCheckLab) Difficulty() Difficulty { return DifficultyHard }
func (l *EtcdHealthCheckLab) EstimatedTime() int     { return 30 }
func (l *EtcdHealthCheckLab) Tags() []string {
	return []string{"etcd", "health", "cluster", "troubleshooting"}
}
func (l *EtcdHealthCheckLab) Cert() Cert        { return CertCKA }
func (l *EtcdHealthCheckLab) DomainWeight() int { return 30 }

func (l *EtcdHealthCheckLab) Description() string {
	return `The etcd cluster is unhealthy. Diagnose the health issue by checking
etcd member list, endpoint health, and fixing any connectivity or
certificate problems.`
}

func (l *EtcdHealthCheckLab) Hints() []string {
	return []string{
		"Check etcd endpoint health",
		"Verify member list",
		"Check certificates and connectivity",
	}
}

func (l *EtcdHealthCheckLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *EtcdHealthCheckLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *EtcdHealthCheckLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "-n", "kube-system",
		"etcd-master", "--", "etcdctl", "endpoint", "health",
		"--endpoints=https://127.0.0.1:2379",
		"--cacert=/etc/kubernetes/pki/etcd/ca.crt",
		"--cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt",
		"--key=/etc/kubernetes/pki/etcd/healthcheck-client.key")
	if err != nil {
		return fmt.Errorf("etcd health check failed: %w", err)
	}
	if strings.Contains(output, "unhealthy") || strings.Contains(output, "error") {
		return fmt.Errorf("etcd is not healthy")
	}
	return nil
}

func (l *EtcdHealthCheckLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check etcd health", Command: "ETCDCTL_API=3 etcdctl endpoint health --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt --key=/etc/kubernetes/pki/etcd/healthcheck-client.key"},
		{Description: "Check member list", Command: "ETCDCTL_API=3 etcdctl member list --endpoints=https://127.0.0.1:2379 --cacert=/etc/kubernetes/pki/etcd/ca.crt --cert=/etc/kubernetes/pki/etcd/healthcheck-client.crt --key=/etc/kubernetes/pki/etcd/healthcheck-client.key"},
		{Description: "Fix certificates", Command: "Renew etcd certificates if expired"},
	}
}
