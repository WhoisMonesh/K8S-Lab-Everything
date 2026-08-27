package labs

import (
	"context"
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
	return `The cluster pulls images directly from Docker Hub. This creates a dependency
on external registries and potential supply chain risks.

Your task: Configure a registry mirror on all worker nodes so that images
from 'docker.io' are pulled from the private mirror 'mirror.example.com' instead.`
}

func (l *CKSRegistryMirrorLab) Hints() []string {
	return []string{
		"Configure /etc/containerd/config.toml on each node",
		"Add a mirror for docker.io registry",
		"Restart containerd after changes",
	}
}

func (l *CKSRegistryMirrorLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSRegistryMirrorLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSRegistryMirrorLab) Verify(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSRegistryMirrorLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Configure containerd mirror", Command: `sudo tee /etc/containerd/config.toml <<EOF
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["https://mirror.example.com"]
EOF`},
		{Description: "Restart containerd", Command: "sudo systemctl restart containerd"},
		{Description: "Test pull from mirror", Command: "crictl pull nginx:latest"},
	}
}
