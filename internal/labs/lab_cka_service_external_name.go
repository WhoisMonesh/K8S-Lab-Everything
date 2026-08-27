package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ServiceExternalNameLab{})
}

type ServiceExternalNameLab struct {
	BaseLab
}

func (l *ServiceExternalNameLab) ID() string { return "cka_service_external_name" }
func (l *ServiceExternalNameLab) Title() string {
	return "Create ExternalName Service"
}
func (l *ServiceExternalNameLab) Category() Category     { return CategoryServicesNetworking }
func (l *ServiceExternalNameLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *ServiceExternalNameLab) EstimatedTime() int     { return 15 }
func (l *ServiceExternalNameLab) Tags() []string {
	return []string{"service", "externalname", "dns", "networking"}
}
func (l *ServiceExternalNameLab) Cert() Cert        { return CertCKA }
func (l *ServiceExternalNameLab) DomainWeight() int { return 20 }

func (l *ServiceExternalNameLab) Description() string {
	return `Create an ExternalName service that aliases an external database host.
The service should be named db-external and point to db.example.com.`
}

func (l *ServiceExternalNameLab) Hints() []string {
	return []string{
		"Set type to ExternalName",
		"Set externalName to the target hostname",
		"Use CNAME DNS record internally",
	}
}

func (l *ServiceExternalNameLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceExternalNameLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *ServiceExternalNameLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "service", "db-external",
		"-n", "external-ns", "-o", "jsonpath={.spec.type}")
	if err != nil {
		return err
	}
	if output != "ExternalName" {
		return fmt.Errorf("service type is not ExternalName")
	}
	return nil
}

func (l *ServiceExternalNameLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create ExternalName service", Command: "cat <<EOF | kubectl apply -f -\napiVersion: v1\nkind: Service\nmetadata:\n  name: db-external\n  namespace: external-ns\nspec:\n  type: ExternalName\n  externalName: db.example.com\nEOF"},
		{Description: "Verify", Command: "kubectl get svc db-external -n external-ns"},
		{Description: "Test DNS", Command: "kubectl exec -n external-ns <pod> -- nslookup db-external.external-ns.svc.cluster.local"},
	}
}
