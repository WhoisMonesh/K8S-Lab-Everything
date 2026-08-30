package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&CKSRegistryMirrorLab{})
}

type CKSRegistryMirrorLab struct {
	BaseLab
}

func (l *CKSRegistryMirrorLab) ID() string             { return "cks_registry_mirror" }
func (l *CKSRegistryMirrorLab) Title() string          { return "Configure Registry Mirror" }
func (l *CKSRegistryMirrorLab) Category() Category     { return CategorySupplyChain }
func (l *CKSRegistryMirrorLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSRegistryMirrorLab) EstimatedTime() int     { return 20 }
func (l *CKSRegistryMirrorLab) Cert() Cert             { return CertCKS }
func (l *CKSRegistryMirrorLab) DomainWeight() int      { return 20 }
func (l *CKSRegistryMirrorLab) Tags() []string {
	return []string{"cks", "registry", "mirror", "supply-chain"}
}

func (l *CKSRegistryMirrorLab) Description() string {
	return `The nodes pull images directly from Docker Hub. There is no registry mirror
configured, creating a dependency on external registries and supply chain risk.

Your task: Configure containerd on the worker (and control-plane) nodes so that
images from 'docker.io' are pulled from the private mirror 'mirror.example.com'
instead.`
}

func (l *CKSRegistryMirrorLab) Hints() []string {
	return []string{
		"Enter the node shell with docker exec",
		"Edit /etc/containerd/config.toml to add a mirror block",
		"Restart containerd after the change",
	}
}

func (l *CKSRegistryMirrorLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSRegistryMirrorLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSRegistryMirrorLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *CKSRegistryMirrorLab) Verify(ctx context.Context, kubeconfigPath string) error {
	worker, err := getWorkerNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	if err := verifyNodeMirror(worker); err != nil {
		return err
	}
	cp, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}
	return verifyNodeMirror(cp)
}

func verifyNodeMirror(node string) error {
	output, err := dockerCommand(node, "cat /etc/containerd/config.toml")
	if err != nil {
		return fmt.Errorf("reading containerd config on node %s: %w", node, err)
	}
	if !strings.Contains(output, `mirrors."docker.io"`) {
		return fmt.Errorf("registry mirror not configured on node %s", node)
	}
	if !strings.Contains(output, "mirror.example.com") {
		return fmt.Errorf("registry mirror not configured on node %s", node)
	}
	return nil
}

func (l *CKSRegistryMirrorLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Enter the worker node shell (kind has no SSH)", Command: "docker exec -it <worker> bash"},
		{Description: "Add the mirror block to containerd config", Command: `cat >> /etc/containerd/config.toml <<EOF
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["https://mirror.example.com"]
EOF`},
		{Description: "Restart containerd", Command: "systemctl restart containerd && exit"},
	}
}
