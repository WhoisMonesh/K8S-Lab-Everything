package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&KubeadmCertRenewal{})
}

type KubeadmCertRenewal struct {
	BaseLab
}

func (l *KubeadmCertRenewal) ID() string             { return "kubeadm_cert_renewal" }
func (l *KubeadmCertRenewal) Title() string          { return "Kubeadm Certificate Expired" }
func (l *KubeadmCertRenewal) Category() Category     { return CategoryControlPlane }
func (l *KubeadmCertRenewal) Difficulty() Difficulty { return DifficultyHard }
func (l *KubeadmCertRenewal) EstimatedTime() int     { return 30 }
func (l *KubeadmCertRenewal) Tags() []string {
	return []string{"certificates", "kubeadm", "control-plane"}
}

func (l *KubeadmCertRenewal) Description() string {
	return `The cluster certificates have expired. kubeadm cannot communicate with the API server.
Renew the certificates using kubeadm.`
}

func (l *KubeadmCertRenewal) Hints() []string {
	return []string{
		"Check certificate expiration",
		"Use kubeadm certs check-expiration",
		"Renew all certificates with kubeadm certs renew all",
	}
}

func (l *KubeadmCertRenewal) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeadmCertRenewal) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *KubeadmCertRenewal) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes")
	if err != nil {
		return err
	}
	if output == "" {
		return fmt.Errorf("cannot get nodes")
	}
	return nil
}

func (l *KubeadmCertRenewal) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check certificates", Command: "sudo kubeadm certs check-expiration"},
		{Description: "Renew certificates", Command: "sudo kubeadm certs renew all"},
		{Description: "Restart control plane", Command: "sudo systemctl restart kubelet"},
		{Description: "Update kubeconfig", Command: "sudo kubeadm kubeconfig user --client-name=ubuntu --org=system:masters > admin.conf"},
	}
}
