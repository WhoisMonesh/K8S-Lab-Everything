package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&EndpointSlicesLab{})
}

type EndpointSlicesLab struct {
	BaseLab
}

func (l *EndpointSlicesLab) ID() string             { return "cka_endpoint_slices" }
func (l *EndpointSlicesLab) Title() string          { return "Configure EndpointSlices" }
func (l *EndpointSlicesLab) Category() Category     { return CategoryServicesNetworking }
func (l *EndpointSlicesLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *EndpointSlicesLab) EstimatedTime() int     { return 20 }
func (l *EndpointSlicesLab) Tags() []string {
	return []string{"endpointslices", "networking", "service"}
}
func (l *EndpointSlicesLab) Cert() Cert        { return CertCKA }
func (l *EndpointSlicesLab) DomainWeight() int { return 20 }

func (l *EndpointSlicesLab) Description() string {
	return `A service has too many endpoints causing performance issues. Configure
EndpointSlices to properly distribute endpoints across multiple slice
resources with a maximum of 50 endpoints per slice.`
}

func (l *EndpointSlicesLab) Hints() []string {
	return []string{
		"Check existing EndpointSlices",
		"Create a custom EndpointSlice with max endpoints per slice",
		"Use the same labels as the service",
	}
}

func (l *EndpointSlicesLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *EndpointSlicesLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *EndpointSlicesLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "endpointslices", "-n", "slice-ns",
		"-o", "name")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "endpointslice") {
		return fmt.Errorf("EndpointSlice not found")
	}
	return nil
}

func (l *EndpointSlicesLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check EndpointSlices", Command: "kubectl get endpointslices -n slice-ns"},
		{Description: "Create EndpointSlice", Command: "cat <<EOF | kubectl apply -f -\napiVersion: discovery.k8s.io/v1\nkind: EndpointSlice\nmetadata:\n  name: my-slice\n  namespace: slice-ns\n  labels:\n    kubernetes.io/service-name: my-service\naddressType: IPv4\nports:\n- port: 80\n  protocol: TCP\nEOF"},
		{Description: "Verify", Command: "kubectl get endpointslices -n slice-ns"},
	}
}
