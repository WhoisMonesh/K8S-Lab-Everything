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

// ClusterSpec declares a single control-plane cluster; a pending node is added
// below so the learner can practice adding a second control plane.
func (l *ControlPlaneHALab) ClusterSpec() ClusterSpec {
	return ClusterSpec{
		Provider:          "kind",
		KubernetesVersion: "v1.28.0",
		Workers:           0,
	}
}

// RequiredNodes provisions a pending control-plane-ready node container.
func (l *ControlPlaneHALab) RequiredNodes() []PendingNode {
	return []PendingNode{{Name: "pending-cp", Version: "v1.28.0"}}
}

func (l *ControlPlaneHALab) Description() string {
	return `Configure the control plane for high availability with a stacked etcd
topology. Add an additional control plane node and configure the load balancer
for API server redundancy.

A pending control-plane-ready node container has been provisioned (named
<cluster>-pending-cp). kind nodes have no SSH — access node shells with:
    docker exec -it <cluster>-control-plane bash
    docker exec -it <cluster>-pending-cp bash`
}

func (l *ControlPlaneHALab) Hints() []string {
	return []string{
		"Enter the control-plane node shell to run kubeadm commands",
		"Upload certs with kubeadm init phase upload-certs --upload-certs",
		"Use kubeadm join with --control-plane flag on the pending node",
	}
}

func (l *ControlPlaneHALab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ControlPlaneHALab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *ControlPlaneHALab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Count control-plane nodes via the real kubeadm role label
	// (node-role.kubernetes.io/control-plane). The filter selects only nodes on
	// which that label key exists; we tally their names.
	count := func(labelKey string) int {
		out, err := kubectl(ctx, kubeconfigPath, "get", "nodes", "-o",
			"jsonpath={range .items[?(@.metadata.labels."+labelKey+")]}{.metadata.name}{\"\\n\"}{end}")
		if err != nil {
			return 0
		}
		return len(strings.Fields(out))
	}
	cp := count("node-role\\.kubernetes\\.io/control-plane") +
		count("node-role\\.kubernetes\\.io/master")
	if cp < 2 {
		return fmt.Errorf("cluster needs at least 2 control plane nodes for HA")
	}
	return nil
}

func (l *ControlPlaneHALab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Enter the control-plane node shell", Command: "docker exec -it <cluster>-control-plane bash"},
		{Description: "Upload certs for the new control plane", Command: "kubeadm init phase upload-certs --upload-certs"},
		{Description: "Generate control plane join command", Command: "kubeadm token create --print-join-command"},
		{Description: "Exit and enter the pending control-plane node", Command: "exit && docker exec -it <cluster>-pending-cp bash"},
		{Description: "Join as a control plane (paste the join command + --control-plane)", Command: "kubeadm join <lb-ip>:6443 --token <token> --discovery-token-ca-cert-hash sha256:<hash> --control-plane --certificate-key <cert-key>"},
		{Description: "Exit and verify HA", Command: "exit && kubectl get nodes"},
	}
}
