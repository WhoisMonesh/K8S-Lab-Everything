package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CertExpirationLab{})
}

type CertExpirationLab struct {
	BaseLab
}

func (l *CertExpirationLab) ID() string {
	return "cert_expiration"
}

func (l *CertExpirationLab) Title() string {
	return "Certificate Expiration Problems"
}

func (l *CertExpirationLab) Category() Category {
	return CategorySecurity
}

func (l *CertExpirationLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *CertExpirationLab) Description() string {
	return `There are concerns about certificate expiration in the cluster.
You need to check the certificate expiration dates for critical components
and document which certificates need renewal.

Your task: Investigate the certificates, identify expiration dates, and
create a ConfigMap documenting the certificate status and renewal needs.`
}

func (l *CertExpirationLab) Hints() []string {
	return []string{
		"Kubernetes certificates are stored in /etc/kubernetes/pki/",
		"Use 'openssl x509' command to check certificate details",
		"The API server certificate is critical to check",
		"Check both the expiration date and the validity period",
	}
}

func (l *CertExpirationLab) EstimatedTime() int {
	return 25
}

func (l *CertExpirationLab) Tags() []string {
	return []string{"certificates", "security", "tls", "pki", "troubleshooting"}
}

func (l *CertExpirationLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CertExpirationLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	containerName := nodeName

	// Create a marker file with certificate information
	certInfo := `Certificate Status Report
========================

Location: /etc/kubernetes/pki/

Certificates to check:
- apiserver.crt (API Server certificate)
- apiserver-kubelet-client.crt (API Server to Kubelet client)
- front-proxy-client.crt (Front proxy client)
- etcd/server.crt (etcd server certificate)

Task: Use openssl to check the expiration dates and document findings.

Common command: openssl x509 -in <cert-file> -noout -dates
`

	cmd := fmt.Sprintf("cat > /tmp/cert-check-required.txt << 'EOF'\n%s\nEOF", certInfo)
	_, err = dockerExec(ctx, containerName, "sh", "-c", cmd)
	if err != nil {
		return fmt.Errorf("creating cert info file: %w", err)
	}

	return nil
}

func (l *CertExpirationLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	containerName := nodeName

	// Check if the marker file exists
	_, err = dockerExec(ctx, containerName, "test", "-f", "/tmp/cert-check-required.txt")
	if err != nil {
		return fmt.Errorf("cert-check marker not found")
	}

	return nil
}

func (l *CertExpirationLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if a certificate-status ConfigMap has been created
	output, err := kubectl(ctx, kubeconfigPath, "get", "configmap",
		"certificate-status", "-n", "kube-system",
		"-o", "jsonpath={.data.check-completed}")
	if err != nil {
		return fmt.Errorf("certificate-status ConfigMap not found: %w", err)
	}

	if !strings.Contains(output, "true") {
		return fmt.Errorf("certificate check not marked as completed")
	}

	// Check if certificate locations are documented
	locations, err := kubectl(ctx, kubeconfigPath, "get", "configmap",
		"certificate-status", "-n", "kube-system",
		"-o", "jsonpath={.data.cert-location}")
	if err == nil && strings.Contains(locations, "/etc/kubernetes/pki") {
		return nil
	}

	return fmt.Errorf("certificate location not properly documented")
}

func (l *CertExpirationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Access the control plane node",
			Command:     "docker exec -it <cluster-name>-control-plane bash",
			Notes:       "You need to access the node to check certificates",
		},
		{
			Description: "Read the certificate check requirements",
			Command:     "cat /tmp/cert-check-required.txt",
			Notes:       "This tells you what needs to be checked",
		},
		{
			Description: "Navigate to the PKI directory",
			Command:     "cd /etc/kubernetes/pki",
			Notes:       "This is where Kubernetes stores its certificates",
		},
		{
			Description: "List the certificates",
			Command:     "ls -la *.crt",
			Notes:       "See all certificate files",
		},
		{
			Description: "Check API server certificate expiration",
			Command:     "openssl x509 -in apiserver.crt -noout -dates",
			Notes:       "This shows the validity period (notBefore and notAfter)",
		},
		{
			Description: "Check other critical certificates",
			Command:     "openssl x509 -in apiserver-kubelet-client.crt -noout -dates",
			Notes:       "Check each important certificate",
		},
		{
			Description: "Check etcd certificate",
			Command:     "openssl x509 -in etcd/server.crt -noout -dates",
			Notes:       "etcd certificates are in a subdirectory",
		},
		{
			Description: "Exit the container and document findings",
			Command: `kubectl create configmap certificate-status -n kube-system \
  --from-literal=check-completed=true \
  --from-literal=cert-location=/etc/kubernetes/pki \
  --from-literal=status="All certificates checked"`,
			Notes: "Create a ConfigMap to document that you've completed the certificate audit",
		},
		{
			Description: "Verify the solution",
			Command:     "kubectl get configmap certificate-status -n kube-system -o yaml",
			Notes:       "Confirm the certificate status has been documented",
		},
	}
}
