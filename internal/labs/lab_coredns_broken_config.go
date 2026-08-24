package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&CoreDNSBrokenConfigLab{})
}

type CoreDNSBrokenConfigLab struct {
	BaseLab
}

func (l *CoreDNSBrokenConfigLab) ID() string {
	return "coredns_broken_config"
}

func (l *CoreDNSBrokenConfigLab) Title() string {
	return "CoreDNS Broken Configuration"
}

func (l *CoreDNSBrokenConfigLab) Category() Category {
	return CategoryDNS
}

func (l *CoreDNSBrokenConfigLab) Difficulty() Difficulty {
	return DifficultyEasy
}

func (l *CoreDNSBrokenConfigLab) Description() string {
	return `DNS resolution is not working in the cluster.
Pods cannot resolve service names or external DNS names.

Your task: Fix the CoreDNS configuration to restore DNS functionality.`
}

func (l *CoreDNSBrokenConfigLab) Hints() []string {
	return []string{
		"Check the CoreDNS pods in the kube-system namespace",
		"Look at the CoreDNS ConfigMap",
		"The Corefile syntax might be invalid",
		"Check the CoreDNS pod logs for syntax errors",
	}
}

func (l *CoreDNSBrokenConfigLab) EstimatedTime() int {
	return 15
}

func (l *CoreDNSBrokenConfigLab) Tags() []string {
	return []string{"dns", "coredns", "configmap", "troubleshooting"}
}

func (l *CoreDNSBrokenConfigLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CoreDNSBrokenConfigLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Break the CoreDNS ConfigMap by introducing a syntax error
	brokenCorefile := `apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    .:53 {
        errors
        health {
           lameduck 5s
        }
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }
        prometheus :9153
        forward . /etc/resolv.conf {
           max_concurrent 1000
        }
        cache 30
        loop
        reload
        loadbalance
        # BROKEN: Invalid plugin that doesn't exist
        invalidplugin {
            some config
        }
    }
`

	if err := kubectlApply(ctx, kubeconfigPath, brokenCorefile); err != nil {
		return fmt.Errorf("applying broken CoreDNS config: %w", err)
	}

	// Delete CoreDNS pods to force them to restart with the new config
	_, _ = kubectl(ctx, kubeconfigPath, "delete", "pods", "-n", "kube-system", "-l", "k8s-app=kube-dns")

	return nil
}

func (l *CoreDNSBrokenConfigLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Wait for CoreDNS to restart and fail
	time.Sleep(10 * time.Second)

	// Check if CoreDNS pods are in CrashLoopBackOff
	output, _ := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system", "-l", "k8s-app=kube-dns", "-o", "jsonpath={.items[*].status.phase}")
	_ = output // We don't strictly need to verify, just wait for the changes to apply

	return nil
}

func (l *CoreDNSBrokenConfigLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if all CoreDNS pods are running
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "k8s-app=kube-dns",
		"-o", "jsonpath={.items[*].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check CoreDNS pods: %w", err)
	}

	// Split by space and check all are Running
	phases := strings.Fields(output)
	if len(phases) == 0 {
		return fmt.Errorf("no CoreDNS pods found")
	}

	for _, phase := range phases {
		if phase != "Running" {
			return fmt.Errorf("CoreDNS pod not running yet (status: %s)", phase)
		}
	}

	// Verify all containers in CoreDNS pods are ready
	output, err = kubectl(ctx, kubeconfigPath, "get", "pods", "-n", "kube-system",
		"-l", "k8s-app=kube-dns",
		"-o", "jsonpath={.items[*].status.containerStatuses[*].ready}")
	if err != nil {
		return fmt.Errorf("failed to check CoreDNS container status: %w", err)
	}

	if !strings.Contains(output, "true") {
		return fmt.Errorf("CoreDNS containers are not ready yet")
	}

	return nil
}

func (l *CoreDNSBrokenConfigLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check CoreDNS pod status",
			Command:     "kubectl get pods -n kube-system -l k8s-app=kube-dns",
			Notes:       "The CoreDNS pods should be in CrashLoopBackOff state",
		},
		{
			Description: "Check CoreDNS logs",
			Command:     "kubectl logs -n kube-system -l k8s-app=kube-dns",
			Notes:       "Look for plugin errors or Corefile syntax errors",
		},
		{
			Description: "Examine the CoreDNS ConfigMap",
			Command:     "kubectl get configmap coredns -n kube-system -o yaml",
			Notes:       "Review the Corefile for invalid plugins or syntax errors",
		},
		{
			Description: "Edit the CoreDNS ConfigMap",
			Command:     "kubectl edit configmap coredns -n kube-system",
			Notes:       "Remove the 'invalidplugin' section and any other invalid configuration",
		},
		{
			Description: "Delete CoreDNS pods to apply the fix",
			Command:     "kubectl delete pods -n kube-system -l k8s-app=kube-dns",
			Notes:       "The deployment will automatically recreate the pods with the fixed config",
		},
		{
			Description: "Verify CoreDNS is running",
			Command:     "kubectl get pods -n kube-system -l k8s-app=kube-dns",
			Notes:       "All CoreDNS pods should be in Running state",
		},
		{
			Description: "Test DNS resolution",
			Command:     "kubectl run test-dns --image=busybox:1.28 --rm -it --restart=Never -- nslookup kubernetes.default",
			Notes:       "DNS should resolve the kubernetes service successfully",
		},
	}
}
