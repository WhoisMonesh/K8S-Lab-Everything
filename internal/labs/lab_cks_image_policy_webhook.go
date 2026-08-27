package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSImagePolicyWebhookLab{})
}

type CKSImagePolicyWebhookLab struct {
	BaseLab
}

func (l *CKSImagePolicyWebhookLab) ID() string             { return "cks_image_policy_webhook" }
func (l *CKSImagePolicyWebhookLab) Title() string          { return "Configure ImagePolicyWebhook" }
func (l *CKSImagePolicyWebhookLab) Category() Category     { return CategorySupplyChain }
func (l *CKSImagePolicyWebhookLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSImagePolicyWebhookLab) EstimatedTime() int     { return 30 }
func (l *CKSImagePolicyWebhookLab) Cert() Cert             { return CertCKS }
func (l *CKSImagePolicyWebhookLab) DomainWeight() int      { return 20 }
func (l *CKSImagePolicyWebhookLab) Tags() []string {
	return []string{"cks", "image-policy", "webhook", "supply-chain"}
}

func (l *CKSImagePolicyWebhookLab) Description() string {
	return `The API server does not validate images before allowing them to run. Any image
from any registry can be deployed.

Your task: Configure the ImagePolicyWebhook admission controller with a
configuration that only allows images tagged with 'latest' or specific
version tags (v1.x, v2.x).`
}

func (l *CKSImagePolicyWebhookLab) Hints() []string {
	return []string{
		"Create an ImagePolicy webhook configuration",
		"Configure allowed image tags",
		"Add the webhook config to API server flags",
	}
}

func (l *CKSImagePolicyWebhookLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSImagePolicyWebhookLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSImagePolicyWebhookLab) Verify(ctx context.Context, kubeconfigPath string) error {
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

func (l *CKSImagePolicyWebhookLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create webhook config", Command: `sudo tee /etc/kubernetes/image-policy-webhook-config.yaml <<EOF
apiVersion: imagepolicy.k8s.io/v1alpha1
kind: ImagePolicy
metadata:
  name: image-policy
spec:
  rejections:
  - matchImages:
    - "docker.io/*"
    - "quay.io/*"
EOF`},
		{Description: "Add to API server", Command: "Edit /etc/kubernetes/manifests/kube-apiserver.yaml and add --enable-admission-plugins=ImagePolicyWebhook --image-policy-webhook-config-file=/etc/kubernetes/image-policy-webhook-config.yaml"},
		{Description: "Restart API server", Command: "sudo mv /etc/kubernetes/manifests/kube-apiserver.yaml /tmp/ && sleep 10 && sudo mv /tmp/kube-apiserver.yaml /etc/kubernetes/manifests/"},
	}
}
