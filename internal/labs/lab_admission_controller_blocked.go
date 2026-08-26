package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&AdmissionControllerBlocked{})
}

type AdmissionControllerBlocked struct {
	BaseLab
}

func (l *AdmissionControllerBlocked) ID() string            { return "admission_controller_blocked" }
func (l *AdmissionControllerBlocked) Title() string         { return "Admission Controller Blocking Pods" }
func (l *AdmissionControllerBlocked) Category() Category    { return CategorySecurity }
func (l *AdmissionControllerBlocked) Difficulty() Difficulty { return DifficultyHard }
func (l *AdmissionControllerBlocked) EstimatedTime() int    { return 20 }
func (l *AdmissionControllerBlocked) Tags() []string        { return []string{"security", "admission", "webhook"} }

func (l *AdmissionControllerBlocked) Description() string {
	return `A MutatingAdmissionWebhook is rejecting all pod creation requests.
Debug and fix the webhook configuration to allow pods to be created.`
}

func (l *AdmissionControllerBlocked) Hints() []string {
	return []string{
		"Check ValidatingWebhookConfiguration",
		"Look at the failurePolicy setting",
		"Consider setting failurePolicy to Ignore",
	}
}

func (l *AdmissionControllerBlocked) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *AdmissionControllerBlocked) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: deny-all-webhook
webhooks:
- name: deny-all.example.com
  clientConfig:
    service:
      name: nonexistent-service
      namespace: default
      path: /deny
  failurePolicy: Fail
  rules:
  - apiGroups: [""]
    apiVersions: ["v1"]
    operations: ["CREATE"]
    resources: ["pods"]
  admissionReviewVersions: ["v1"]`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *AdmissionControllerBlocked) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "validatingwebhookconfiguration", "deny-all-webhook",
		"-o", "jsonpath={.webhooks[0].failurePolicy}")
	if err != nil {
		return err
	}
	if output == "Fail" {
		return fmt.Errorf("webhook still has failurePolicy: Fail")
	}
	return nil
}

func (l *AdmissionControllerBlocked) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check webhook config", Command: "kubectl get validatingwebhookconfiguration deny-all-webhook -o yaml"},
		{Description: "Fix failurePolicy", Command: "kubectl patch validatingwebhookconfiguration deny-all-webhook --type='json' -p='[{\"op\": \"replace\", \"path\": \"/webhooks/0/failurePolicy\", \"value\": \"Ignore\"}]'"},
	}
}
