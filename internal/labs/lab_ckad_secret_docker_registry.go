package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADSecretDockerRegistryLab{})
}

type CKADSecretDockerRegistryLab struct {
	BaseLab
}

func (l *CKADSecretDockerRegistryLab) ID() string {
	return "ckad_secret_docker_registry"
}

func (l *CKADSecretDockerRegistryLab) Title() string {
	return "Create Docker Registry Secret"
}

func (l *CKADSecretDockerRegistryLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADSecretDockerRegistryLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADSecretDockerRegistryLab) Cert() Cert             { return CertCKAD }
func (l *CKADSecretDockerRegistryLab) DomainWeight() int      { return 25 }
func (l *CKADSecretDockerRegistryLab) EstimatedTime() int     { return 15 }
func (l *CKADSecretDockerRegistryLab) Tags() []string {
	return []string{"secret", "docker-registry", "image-pull"}
}

func (l *CKADSecretDockerRegistryLab) Description() string {
	return `A pod needs to pull images from a private Docker registry. Create a
Docker registry Secret to authenticate with the registry.

Your task: Create the docker-registry secret for image pulling.`
}

func (l *CKADSecretDockerRegistryLab) Hints() []string {
	return []string{
		"Use kubectl create secret docker-registry",
		"Provide --docker-server, --docker-username, --docker-password",
		"The secret type must be kubernetes.io/dockerconfigjson",
	}
}

func (l *CKADSecretDockerRegistryLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADSecretDockerRegistryLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADSecretDockerRegistryLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "secret", "registry-cred",
		"-o", "jsonpath={.type}")
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}
	if strings.TrimSpace(output) != "kubernetes.io/dockerconfigjson" {
		return fmt.Errorf("secret type is not docker-registry (current: %s)", output)
	}
	return nil
}

func (l *CKADSecretDockerRegistryLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create docker registry secret", Command: "kubectl create secret docker-registry registry-cred --docker-server=registry.example.com --docker-username=user --docker-password=pass"},
		{Description: "Verify", Command: "kubectl get secret registry-cred -o yaml"},
	}
}
