package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADSecurityContextReadOnlyRootLab{})
}

type CKADSecurityContextReadOnlyRootLab struct {
	BaseLab
}

func (l *CKADSecurityContextReadOnlyRootLab) ID() string {
	return "ckad_security_context_readonly_root"
}

func (l *CKADSecurityContextReadOnlyRootLab) Title() string {
	return "Use readOnlyRootFilesystem"
}

func (l *CKADSecurityContextReadOnlyRootLab) Category() Category     { return CategoryAppConfigSecurity }
func (l *CKADSecurityContextReadOnlyRootLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADSecurityContextReadOnlyRootLab) Cert() Cert             { return CertCKAD }
func (l *CKADSecurityContextReadOnlyRootLab) DomainWeight() int      { return 25 }
func (l *CKADSecurityContextReadOnlyRootLab) EstimatedTime() int     { return 15 }
func (l *CKADSecurityContextReadOnlyRootLab) Tags() []string {
	return []string{"read-only-root", "security", "filesystem"}
}

func (l *CKADSecurityContextReadOnlyRootLab) Description() string {
	return `For security hardening, a container should use a read-only root filesystem.
Configure the container to use readOnlyRootFilesystem and use an EmptyDir
volume for temporary writes.

Your task: Add readOnlyRootFilesystem and configure writable volumes.`
}

func (l *CKADSecurityContextReadOnlyRootLab) Hints() []string {
	return []string{
		"Set readOnlyRootFilesystem: true in container securityContext",
		"Add EmptyDir volumes for /tmp and /var/cache/nginx",
		"Mount volumes at paths the application writes to",
	}
}

func (l *CKADSecurityContextReadOnlyRootLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADSecurityContextReadOnlyRootLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: readonly-app
  labels:
    app: readonly-app
spec:
  containers:
  - name: web
    image: nginx:alpine
    ports:
    - containerPort: 80`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADSecurityContextReadOnlyRootLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "readonly-app",
		"-o", "jsonpath={.spec.containers[0].securityContext.readOnlyRootFilesystem}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) != "true" {
		return fmt.Errorf("readOnlyRootFilesystem not set to true (current: %s)", output)
	}
	return nil
}

func (l *CKADSecurityContextReadOnlyRootLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Edit pod", Command: "kubectl edit pod readonly-app"},
		{Description: "Add readOnlyRootFilesystem", Command: "Add securityContext.readOnlyRootFilesystem: true"},
		{Description: "Add writable volumes", Command: "Add EmptyDir volumes mounted at /tmp and /var/cache/nginx"},
		{Description: "Verify", Command: "kubectl get pod readonly-app -o yaml | grep readOnlyRootFilesystem"},
	}
}
