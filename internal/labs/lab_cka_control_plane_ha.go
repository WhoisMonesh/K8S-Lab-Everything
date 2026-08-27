package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&ControlPlaneHALab{})
}

type ControlPlaneHALab struct {
	BaseLab
}

func (l *ControlPlaneHALab) ID() string             { return "cka_control_plane_ha" }
func (l *ControlPlaneHALab) Title() string          { return "Setup High Availability Control Plane" }
func (l *ControlPlaneHALab) Category() Category     { return CategoryClusterArchitecture }
func (l *ControlPlaneHALab) Difficulty() Difficulty { return DifficultyHard }
func (l *ControlPlaneHALab) EstimatedTime() int     { return 45 }
func (l *ControlPlaneHALab) Tags() []string {
	return []string{"ha", "control-plane", "stacked", "etcd"}
}
func (l *ControlPlaneHALab) Cert() Cert        { return CertCKA }
func (l *ControlPlaneHALab) DomainWeight() int { return 25 }

func (l *ControlPlaneHALab) Description() string {
	return `Configure the control plane for high availability with a stacked etcd
topology. Add an additional control plane node and configure the load balancer
for API server redundancy.`
}

func (l *ControlPlaneHALab) Hints() []string {
	return []string{
		"Set up a load balancer for API server",
		"Use kubeadm join with --control-plane flag",
		"Ensure certificates are copied to new node",
	}
}

func (l *ControlPlaneHALab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ControlPlaneHALab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *ControlPlaneHALab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o",
		"jsonpath={.items[*].metadata.labels.node\\.kubernetes\\.io/role}")
	if err != nil {
		return err
	}
	masterCount := strings.Count(output, "master")
	if masterCount < 2 {
		return fmt.Errorf("cluster needs at least 2 control plane nodes for HA")
	}
	return nil
}

func (l *ControlPlaneHALab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Generate control plane join command", Command: "sudo kubeadm token create --print-join-command"},
		{Description: "Join as control plane", Command: "sudo kubeadm join <lb-ip>:6443 --token <token> --discovery-token-ca-cert-hash sha256:<hash> --control-plane --certificate-key <cert-key>"},
		{Description: "Upload certs first", Command: "sudo kubeadm init phase upload-certs --upload-certs"},
		{Description: "Verify HA", Command: "kubectl get nodes -l node-role.kubernetes.io/master"},
	}
}
