package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADConfigMapCreateLab{})
}

type CKADConfigMapCreateLab struct {
	BaseLab
}

func (l *CKADConfigMapCreateLab) ID() string             { return "ckad_configmap_create" }
func (l *CKADConfigMapCreateLab) Title() string          { return "Create ConfigMap from Literals" }
func (l *CKADConfigMapCreateLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADConfigMapCreateLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADConfigMapCreateLab) Cert() Cert             { return CertCKAD }
func (l *CKADConfigMapCreateLab) DomainWeight() int      { return 25 }
func (l *CKADConfigMapCreateLab) EstimatedTime() int     { return 10 }
func (l *CKADConfigMapCreateLab) Tags() []string {
	return []string{"configmap", "configuration", "literals"}
}

func (l *CKADConfigMapCreateLab) Description() string {
	return `An application needs configuration data stored in a ConfigMap.
Create a ConfigMap named 'app-config' with key-value pairs from literals.

Your task: Create the ConfigMap with the required configuration.`
}

func (l *CKADConfigMapCreateLab) Hints() []string {
	return []string{
		"Use kubectl create configmap with --from-literal",
		"Each --from-literal adds a key=value pair",
		"Verify with kubectl get configmap",
	}
}

func (l *CKADConfigMapCreateLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADConfigMapCreateLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADConfigMapCreateLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap", "app-config",
		"-o", "jsonpath={.data}")
	if err != nil {
		return fmt.Errorf("failed to get configmap: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("configmap not found or has no data")
	}
	return nil
}

func (l *CKADConfigMapCreateLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create ConfigMap", Command: "kubectl create configmap app-config --from-literal=database_host=mysql --from-literal=database_port=3306"},
		{Description: "Verify", Command: "kubectl get configmap app-config -o yaml"},
	}
}
