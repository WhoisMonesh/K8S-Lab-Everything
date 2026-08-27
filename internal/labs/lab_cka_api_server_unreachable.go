package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&APIServerUnreachableLab{})
}

type APIServerUnreachableLab struct {
	BaseLab
}

func (l *APIServerUnreachableLab) ID() string { return "cka_api_server_unreachable" }
func (l *APIServerUnreachableLab) Title() string {
	return "Debug API Server Connectivity"
}
func (l *APIServerUnreachableLab) Category() Category     { return CategoryTroubleshooting }
func (l *APIServerUnreachableLab) Difficulty() Difficulty { return DifficultyHard }
func (l *APIServerUnreachableLab) EstimatedTime() int     { return 30 }
func (l *APIServerUnreachableLab) Tags() []string {
	return []string{"api-server", "connectivity", "troubleshooting"}
}
func (l *APIServerUnreachableLab) Cert() Cert        { return CertCKA }
func (l *APIServerUnreachableLab) DomainWeight() int { return 30 }

func (l *APIServerUnreachableLab) Description() string {
	return `The API server is unreachable from a worker node. Diagnose the
connectivity issue by checking certificates, network connectivity,
and API server status.`
}

func (l *APIServerUnreachableLab) Hints() []string {
	return []string{
		"Check API server pod status",
		"Verify certificates are valid",
		"Test network connectivity with curl",
	}
}

func (l *APIServerUnreachableLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *APIServerUnreachableLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *APIServerUnreachableLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-apiserver", "-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return err
	}
	if output != "Running" {
		return fmt.Errorf("API server not running")
	}
	return nil
}

func (l *APIServerUnreachableLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check API server pods", Command: "kubectl get pods -n kube-system -l component=kube-apiserver"},
		{Description: "Check API server logs", Command: "kubectl logs -n kube-system -l component=kube-apiserver --tail=50"},
		{Description: "Verify certificates", Command: "openssl x509 -in /etc/kubernetes/pki/apiserver.crt -noout -dates"},
		{Description: "Test connectivity", Command: "curl -k https://<api-server>:6443/healthz"},
	}
}
