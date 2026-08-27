package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CoreDNSConfigLab{})
}

type CoreDNSConfigLab struct {
	BaseLab
}

func (l *CoreDNSConfigLab) ID() string             { return "cka_coredns_config" }
func (l *CoreDNSConfigLab) Title() string          { return "Customize CoreDNS Configuration" }
func (l *CoreDNSConfigLab) Category() Category     { return CategoryServicesNetworking }
func (l *CoreDNSConfigLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CoreDNSConfigLab) EstimatedTime() int     { return 20 }
func (l *CoreDNSConfigLab) Tags() []string {
	return []string{"coredns", "dns", "configuration"}
}
func (l *CoreDNSConfigLab) Cert() Cert        { return CertCKA }
func (l *CoreDNSConfigLab) DomainWeight() int { return 20 }

func (l *CoreDNSConfigLab) Description() string {
	return `The CoreDNS configuration needs to be customized to add a custom DNS
entry for an external service. Edit the CoreDNS ConfigMap to add a
hosts plugin entry.`
}

func (l *CoreDNSConfigLab) Hints() []string {
	return []string{
		"Edit the coredns ConfigMap in kube-system",
		"Add a hosts block with custom entries",
		"Restart CoreDNS after changes",
	}
}

func (l *CoreDNSConfigLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CoreDNSConfigLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CoreDNSConfigLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap", "coredns",
		"-n", "kube-system", "-o", "jsonpath={.data.Corefile}")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "hosts") {
		return fmt.Errorf("hosts plugin not configured in CoreDNS")
	}
	return nil
}

func (l *CoreDNSConfigLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check CoreDNS config", Command: "kubectl get configmap coredns -n kube-system -o yaml"},
		{Description: "Edit ConfigMap", Command: "kubectl edit configmap coredns -n kube-system"},
		{Description: "Add hosts plugin", Command: "Add hosts block:\nhosts {\n  192.168.1.100 myapp.local\n  fallthrough\n}"},
		{Description: "Restart CoreDNS", Command: "kubectl rollout restart deployment coredns -n kube-system"},
	}
}
