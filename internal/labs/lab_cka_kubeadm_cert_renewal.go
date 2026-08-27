package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&KubeadmCertRenewalLab{})
}

type KubeadmCertRenewalLab struct {
	BaseLab
}

func (l *KubeadmCertRenewalLab) ID() string { return "cka_kubeadm_cert_renewal" }
func (l *KubeadmCertRenewalLab) Title() string {
	return "Renew Expired kubeadm Certificates"
}
func (l *KubeadmCertRenewalLab) Category() Category     { return CategoryClusterArchitecture }
func (l *KubeadmCertRenewalLab) Difficulty() Difficulty { return DifficultyHard }
func (l *KubeadmCertRenewalLab) EstimatedTime() int     { return 25 }
func (l *KubeadmCertRenewalLab) Tags() []string {
	return []string{"kubeadm", "certificates", "tls", "renewal"}
}
func (l *KubeadmCertRenewalLab) Cert() Cert        { return CertCKA }
func (l *KubeadmCertRenewalLab) DomainWeight() int { return 25 }

func (l *KubeadmCertRenewalLab) Description() string {
	return `The kubeadm certificates have expired and the API server is rejecting
connections. Renew all certificates and restart the control plane components.`
}

func (l *KubeadmCertRenewalLab) Hints() []string {
	return []string{
		"Check certificate expiration with kubeadm certs check-expiration",
		"Use kubeadm certs renew all to renew certificates",
		"Restart control plane pods after renewal",
	}
}

func (l *KubeadmCertRenewalLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeadmCertRenewalLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *KubeadmCertRenewalLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "-n", "kube-system",
		"etcd-master", "--", "openssl", "x509", "-in", "/etc/kubernetes/pki/apiserver.crt",
		"-noout", "-dates")
	if err != nil {
		return err
	}
	if strings.Contains(output, "Not After") {
		return nil
	}
	return fmt.Errorf("certificate check failed")
}

func (l *KubeadmCertRenewalLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check expiration", Command: "sudo kubeadm certs check-expiration"},
		{Description: "Renew all certs", Command: "sudo kubeadm certs renew all"},
		{Description: "Restart control plane", Command: "sudo crictl pods --name kube-apiserver -q | xargs sudo crictl stopp"},
		{Description: "Copy new kubeconfig", Command: "sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config"},
	}
}
