package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSRegistryPrivateOnlyLab{})
}

type CKSRegistryPrivateOnlyLab struct {
	BaseLab
}

func (l *CKSRegistryPrivateOnlyLab) ID() string             { return "cks_registry_private_only" }
func (l *CKSRegistryPrivateOnlyLab) Title() string          { return "Restrict to Private Registries" }
func (l *CKSRegistryPrivateOnlyLab) Category() Category     { return CategorySupplyChain }
func (l *CKSRegistryPrivateOnlyLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSRegistryPrivateOnlyLab) EstimatedTime() int     { return 25 }
func (l *CKSRegistryPrivateOnlyLab) Cert() Cert             { return CertCKS }
func (l *CKSRegistryPrivateOnlyLab) DomainWeight() int      { return 20 }
func (l *CKSRegistryPrivateOnlyLab) Tags() []string {
	return []string{"cks", "registry", "image-policy", "supply-chain"}
}

func (l *CKSRegistryPrivateOnlyLab) Description() string {
	return `Pods in the cluster can pull images from any registry including public ones
like Docker Hub. This increases the risk of using untrusted images.

Your task: Configure an ImagePolicyWebhook that only allows images from
the private registry 'registry.example.com'.`
}

func (l *CKSRegistryPrivateOnlyLab) Hints() []string {
	return []string{
		"Create an ImagePolicy configuration",
		"Configure the API server to use --image-policy-webhook-config-file",
		"Ensure the webhook server validates image sources",
	}
}

func (l *CKSRegistryPrivateOnlyLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSRegistryPrivateOnlyLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSRegistryPrivateOnlyLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-apiserver", "-o", "jsonpath={.items[0].spec.containers[0].args}")
	if err != nil {
		return fmt.Errorf("failed to get apiserver args: %w", err)
	}
	if strings.Contains(output, "image-policy-webhook") {
		return nil
	}
	return fmt.Errorf("image policy webhook not configured")
}

func (l *CKSRegistryPrivateOnlyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create ImagePolicy config", Command: "sudo tee /etc/kubernetes/image-policy-webhook-config.json <<EOF\n{\n  \"imagePolicy\": {\n    \"kubeReference\": \"imagePolicy\",\n    \"evaluationMode\": \"alwaysEnforce\",\n    \"batch-size\": 5\n  }\n}\nEOF"},
		{Description: "Add webhook flag to API server", Command: "Edit /etc/kubernetes/manifests/kube-apiserver.yaml and add --image-policy-webhook-config-file=/etc/kubernetes/image-policy-webhook-config.json"},
		{Description: "Restart API server", Command: "sudo mv /etc/kubernetes/manifests/kube-apiserver.yaml /tmp/ && sleep 10 && sudo mv /tmp/kube-apiserver.yaml /etc/kubernetes/manifests/"},
	}
}
