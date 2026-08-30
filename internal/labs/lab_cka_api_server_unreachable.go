package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
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
	return `The API server is unreachable from a worker node. An invalid flag has
been added to the kube-apiserver static pod manifest causing it to crash.
Diagnose and fix the API server configuration.

kind nodes are containers (no SSH); access the control-plane node shell with:
    docker exec -it <cluster>-control-plane bash`
}

func (l *APIServerUnreachableLab) Hints() []string {
	return []string{
		"Check API server pod status",
		"Inspect the apiserver static pod manifest on the control-plane node",
		"Look for invalid flags in the manifest",
	}
}

func (l *APIServerUnreachableLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *APIServerUnreachableLab) Break(ctx context.Context, kubeconfigPath string) error {
	nodeName, err := getControlPlaneNode(ctx, kubeconfigPath)
	if err != nil {
		return err
	}

	output, err := dockerExec(ctx, nodeName, "cat", "/etc/kubernetes/manifests/kube-apiserver.yaml")
	if err != nil {
		return fmt.Errorf("reading kube-apiserver manifest: %w", err)
	}

	modifiedManifest := strings.Replace(output,
		"- kube-apiserver",
		"- kube-apiserver\n    - --invalid-apiserver-flag=true",
		1)

	writeCmd := fmt.Sprintf("cat > /etc/kubernetes/manifests/kube-apiserver.yaml << 'EOF'\n%s\nEOF", modifiedManifest)
	_, err = dockerExec(ctx, nodeName, "sh", "-c", writeCmd)
	if err != nil {
		return fmt.Errorf("writing modified manifest: %w", err)
	}

	return nil
}

func (l *APIServerUnreachableLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *APIServerUnreachableLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "component=kube-apiserver", "-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return err
	}
	if !strings.Contains(output, "Running") {
		return fmt.Errorf("API server not running")
	}
	return nil
}

func (l *APIServerUnreachableLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check API server pod status", Command: "kubectl get pods -n kube-system | grep apiserver"},
		{Description: "Access the control plane node", Command: "docker exec -it <cluster>-control-plane bash"},
		{Description: "Examine the apiserver manifest", Command: "cat /etc/kubernetes/manifests/kube-apiserver.yaml"},
		{Description: "Remove the invalid flag", Command: "sed -i '/--invalid-apiserver-flag/d' /etc/kubernetes/manifests/kube-apiserver.yaml"},
		{Description: "Verify API server is running", Command: "kubectl get pods -n kube-system | grep apiserver"},
	}
}
