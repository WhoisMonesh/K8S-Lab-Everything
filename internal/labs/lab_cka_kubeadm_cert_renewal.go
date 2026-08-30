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
	return `The kubeadm certificates are nearing/at expiry and the control plane is
rejecting connections. Renew all certificates and restart the control plane
static pods.

kind nodes are containers (no SSH); run kubeadm inside the control-plane node:
    docker exec -it <cluster>-control-plane bash`
}

func (l *KubeadmCertRenewalLab) Hints() []string {
	return []string{
		"Enter the control-plane node: docker exec -it <cluster>-control-plane bash",
		"Check certificate expiration with kubeadm certs check-expiration",
		"Use kubeadm certs renew all to renew certificates",
		"Restart control plane static pods after renewal",
	}
}

func (l *KubeadmCertRenewalLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *KubeadmCertRenewalLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("getting control plane node: %w", err)
	}
	if _, err := dockerCommand(nodeName, "cp /etc/kubernetes/pki/apiserver.crt /etc/kubernetes/pki/apiserver.crt.bak"); err != nil {
		return fmt.Errorf("backing up cert: %w", err)
	}
	if _, err := dockerCommand(nodeName, "truncate -s 10 /etc/kubernetes/pki/apiserver.crt"); err != nil {
		return fmt.Errorf("truncating cert: %w", err)
	}
	return nil
}

func (l *KubeadmCertRenewalLab) Verify(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	output, err := dockerCommand(nodeName, "openssl x509 -in /etc/kubernetes/pki/apiserver.crt -noout -checkend 86400")
	if err != nil || strings.Contains(output, "will expire") || strings.Contains(output, "error") {
		return fmt.Errorf("certificate is still expired or invalid - renew with kubeadm certs renew all")
	}
	return nil
}

func (l *KubeadmCertRenewalLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Enter the control-plane node", Command: "docker exec -it <cluster>-control-plane bash", Notes: "All following commands run inside the container"},
		{Description: "Export kubeadm config from ConfigMap", Command: "kubectl -n kube-system get cm kubeadm-config -o jsonpath='{.data.ClusterConfiguration}' > /etc/kubernetes/kubeadm-config.yaml", Notes: "Required for cert regeneration"},
		{Description: "Remove corrupted certs", Command: "cd /etc/kubernetes/pki && rm -f apiserver.crt apiserver.key apiserver-kubelet-client.crt apiserver-kubelet-client.key front-proxy-ca.crt front-proxy-ca.key front-proxy-client.crt front-proxy-client.key sa.key sa.pub", Notes: "Clear corrupted certs"},
		{Description: "Remove old kubeconfigs", Command: "rm -f /etc/kubernetes/admin.conf /etc/kubernetes/kubelet.conf /etc/kubernetes/controller-manager.conf /etc/kubernetes/scheduler.conf", Notes: "Must regenerate with new CA"},
		{Description: "Regenerate all certificates", Command: "kubeadm init phase certs all --config /etc/kubernetes/kubeadm-config.yaml", Notes: "Creates fresh CA + all signed certs"},
		{Description: "Regenerate all kubeconfigs", Command: "kubeadm init phase kubeconfig all --config /etc/kubernetes/kubeadm-config.yaml", Notes: "Creates admin.conf, kubelet.conf, etc."},
		{Description: "Exit container", Command: "exit", Notes: "Return to host for remaining steps"},
		{Description: "Refresh host kubeconfig with kind", Command: "kind export kubeconfig --name <cluster-name>", Notes: "Updates ~/.kube/config with correct server address (127.0.0.1:6443) and new CA"},
		{Description: "Verify cluster", Command: "kubectl get nodes && kubectl get pods -A", Notes: "Cluster should be healthy"},
	}
}
