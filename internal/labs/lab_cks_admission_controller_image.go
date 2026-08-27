package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSAdmissionControllerImageLab{})
}

type CKSAdmissionControllerImageLab struct {
	BaseLab
}

func (l *CKSAdmissionControllerImageLab) ID() string             { return "cks_admission_controller_image" }
func (l *CKSAdmissionControllerImageLab) Title() string          { return "Validate Images in Admission" }
func (l *CKSAdmissionControllerImageLab) Category() Category     { return CategorySupplyChain }
func (l *CKSAdmissionControllerImageLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSAdmissionControllerImageLab) EstimatedTime() int     { return 30 }
func (l *CKSAdmissionControllerImageLab) Cert() Cert             { return CertCKS }
func (l *CKSAdmissionControllerImageLab) DomainWeight() int      { return 20 }
func (l *CKSAdmissionControllerImageLab) Tags() []string {
	return []string{"cks", "admission", "image-validation", "supply-chain"}
}

func (l *CKSAdmissionControllerImageLab) Description() string {
	return `A validating webhook should be configured to check that all container images
come from the approved registry 'registry.example.com'. Pods using images
from other registries should be rejected.

Your task: Create a ValidatingWebhookConfiguration that intercepts pod creation
and validates the image registry source.`
}

func (l *CKSAdmissionControllerImageLab) Hints() []string {
	return []string{
		"Create a ValidatingWebhookConfiguration",
		"Configure rules to intercept Pod CREATE operations",
		"Reference a webhook service that validates images",
	}
}

func (l *CKSAdmissionControllerImageLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSAdmissionControllerImageLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSAdmissionControllerImageLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "validatingwebhookconfigurations", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}
	if strings.Contains(output, "image-validation") || strings.Contains(output, "image-check") {
		return nil
	}
	return fmt.Errorf("image validation webhook not found")
}

func (l *CKSAdmissionControllerImageLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create webhook configuration", Command: `cat <<EOF | kubectl apply -f -
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: image-validation
webhooks:
- name: validate-image.registry.example.com
  rules:
  - apiGroups: [""]
    apiVersions: ["v1"]
    operations: ["CREATE"]
    resources: ["pods"]
  clientConfig:
    service:
      name: image-validator
      namespace: kube-system
      path: /validate
    caBundle: <base64-ca-cert>
  admissionReviewVersions: ["v1"]
EOF`},
		{Description: "Verify", Command: "kubectl get validatingwebhookconfigurations"},
	}
}
