package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADConfigMapFileLab{})
}

type CKADConfigMapFileLab struct {
	BaseLab
}

func (l *CKADConfigMapFileLab) ID() string             { return "ckad_configmap_file" }
func (l *CKADConfigMapFileLab) Title() string          { return "Create ConfigMap from File" }
func (l *CKADConfigMapFileLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADConfigMapFileLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADConfigMapFileLab) Cert() Cert             { return CertCKAD }
func (l *CKADConfigMapFileLab) DomainWeight() int      { return 25 }
func (l *CKADConfigMapFileLab) EstimatedTime() int     { return 10 }
func (l *CKADConfigMapFileLab) Tags() []string {
	return []string{"configmap", "file", "configuration"}
}

func (l *CKADConfigMapFileLab) Description() string {
	return `An application needs configuration from a file. Create a ConfigMap named
'file-config' from a file containing key-value pairs.

Your task: Create the ConfigMap from the configuration file.`
}

func (l *CKADConfigMapFileLab) Hints() []string {
	return []string{
		"Use kubectl create configmap with --from-file",
		"The file name becomes the key by default",
		"You can specify a custom key with key=file syntax",
	}
}

func (l *CKADConfigMapFileLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADConfigMapFileLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKADConfigMapFileLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap", "file-config",
		"-o", "jsonpath={.data}")
	if err != nil {
		return fmt.Errorf("failed to get configmap: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("configmap not found or has no data")
	}
	return nil
}

func (l *CKADConfigMapFileLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create config file", Command: "echo 'key1=value1\nkey2=value2' > config.txt"},
		{Description: "Create ConfigMap", Command: "kubectl create configmap file-config --from-file=config.txt"},
		{Description: "Verify", Command: "kubectl get configmap file-config -o yaml"},
	}
}
