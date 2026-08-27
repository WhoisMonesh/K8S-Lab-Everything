package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSAPIServerAdmissionLab{})
}

type CKSAPIServerAdmissionLab struct {
	BaseLab
}

func (l *CKSAPIServerAdmissionLab) ID() string             { return "cks_api_server_admission" }
func (l *CKSAPIServerAdmissionLab) Title() string          { return "Configure Admission Controllers" }
func (l *CKSAPIServerAdmissionLab) Category() Category     { return CategoryClusterHardening }
func (l *CKSAPIServerAdmissionLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSAPIServerAdmissionLab) EstimatedTime() int     { return 30 }
func (l *CKSAPIServerAdmissionLab) Cert() Cert             { return CertCKS }
func (l *CKSAPIServerAdmissionLab) DomainWeight() int      { return 15 }
func (l *CKSAPIServerAdmissionLab) Tags() []string {
	return []string{"cks", "admission", "api-server", "security"}
}

func (l *CKSAPIServerAdmissionLab) Description() string {
	return `The API server does not have the NodeRestriction admission controller enabled.
This controller limits the Node and Pod resources a kubelet can modify.

Your task: Enable the NodeRestriction admission controller by adding it to the
--enable-admission-plugins flag in the kube-apiserver manifest.`
}

func (l *CKSAPIServerAdmissionLab) Hints() []string {
	return []string{
		"Check the kube-apiserver manifest for admission plugins",
		"Add NodeRestriction to the --enable-admission-plugins flag",
		"The API server pod will restart automatically",
	}
}

func (l *CKSAPIServerAdmissionLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSAPIServerAdmissionLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSAPIServerAdmissionLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-apiserver", "-o", "jsonpath={.items[0].spec.containers[0].args}")
	if err != nil {
		return fmt.Errorf("failed to get apiserver args: %w", err)
	}
	if strings.Contains(output, "NodeRestriction") {
		return nil
	}
	return fmt.Errorf("NodeRestriction admission controller not enabled")
}

func (l *CKSAPIServerAdmissionLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current admission plugins", Command: "cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep enable-admission-plugins"},
		{Description: "Add NodeRestriction plugin", Command: "sudo sed -i 's/--enable-admission-plugins=NodeRestriction/--enable-admission-plugins=NodeRestriction,PodSecurity/' /etc/kubernetes/manifests/kube-apiserver.yaml"},
		{Description: "Verify API server restart", Command: "kubectl get pods -n kube-system -l component=kube-apiserver"},
	}
}
