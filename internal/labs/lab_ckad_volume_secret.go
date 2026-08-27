package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADVolumeSecretLab{})
}

type CKADVolumeSecretLab struct {
	BaseLab
}

func (l *CKADVolumeSecretLab) ID() string             { return "ckad_volume_secret" }
func (l *CKADVolumeSecretLab) Title() string          { return "Mount Secret as Volume" }
func (l *CKADVolumeSecretLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADVolumeSecretLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKADVolumeSecretLab) Cert() Cert             { return CertCKAD }
func (l *CKADVolumeSecretLab) DomainWeight() int      { return 20 }
func (l *CKADVolumeSecretLab) EstimatedTime() int     { return 15 }
func (l *CKADVolumeSecretLab) Tags() []string {
	return []string{"secret", "volumes", "security"}
}

func (l *CKADVolumeSecretLab) Description() string {
	return `An application needs to read TLS certificates from a Secret mounted as a volume.
The Secret exists but is not properly mounted to the pod.

Your task: Mount the Secret 'tls-cert' as a volume in the pod.`
}

func (l *CKADVolumeSecretLab) Hints() []string {
	return []string{
		"Use volumes with secret type",
		"Mount the volume at /etc/tls or similar path",
		"Secret keys become file names in the mount",
	}
}

func (l *CKADVolumeSecretLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADVolumeSecretLab) Break(ctx context.Context, kubeconfigPath string) error {
	secret := `apiVersion: v1
kind: Secret
metadata:
  name: tls-cert
type: Opaque
data:
  tls.crt: Y2VydGlmaWNhdGU=
  tls.key: a2V5`
	if err := kubectlApply(ctx, kubeconfigPath, secret); err != nil {
		return fmt.Errorf("creating secret: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: web-secure
  labels:
    app: web-secure
spec:
  containers:
  - name: web
    image: nginx:alpine
    volumeMounts: []`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKADVolumeSecretLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pod", "web-secure",
		"-o", "jsonpath={.spec.volumes[*].secret.secretName}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no Secret volume found")
	}
	if !strings.Contains(output, "tls-cert") {
		return fmt.Errorf("Secret 'tls-cert' not mounted")
	}
	return nil
}

func (l *CKADVolumeSecretLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check Secret", Command: "kubectl get secret tls-cert -o yaml"},
		{Description: "Edit pod to mount Secret", Command: "kubectl edit pod web-secure"},
		{Description: "Add volume and volumeMount", Command: "Add secret volume and mount at /etc/tls"},
		{Description: "Verify mount", Command: "kubectl exec web-secure -- ls /etc/tls"},
	}
}
