package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&CKSNodeMetadataProtectionLab{})
}

type CKSNodeMetadataProtectionLab struct {
	BaseLab
}

func (l *CKSNodeMetadataProtectionLab) ID() string             { return "cks_node_metadata_protection" }
func (l *CKSNodeMetadataProtectionLab) Title() string          { return "Protect Node Metadata from Pods" }
func (l *CKSNodeMetadataProtectionLab) Category() Category     { return CategoryClusterSetupCKS }
func (l *CKSNodeMetadataProtectionLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSNodeMetadataProtectionLab) EstimatedTime() int     { return 20 }
func (l *CKSNodeMetadataProtectionLab) Cert() Cert             { return CertCKS }
func (l *CKSNodeMetadataProtectionLab) DomainWeight() int      { return 15 }
func (l *CKSNodeMetadataProtectionLab) Tags() []string {
	return []string{"cks", "node-metadata", "network-policy", "security"}
}

func (l *CKSNodeMetadataProtectionLab) Description() string {
	return `Pods in the cluster can access the node metadata endpoint (169.254.169.254)
which exposes sensitive instance information including cloud credentials.

Your task: Create a NetworkPolicy in the 'metadata-protect' namespace that
blocks all egress to the link-local address range 169.254.0.0/16.`
}

func (l *CKSNodeMetadataProtectionLab) Hints() []string {
	return []string{
		"Use ipBlock in the NetworkPolicy egress rule",
		"Block 169.254.0.0/16 to prevent metadata access",
		"Apply the policy to all pods in the namespace",
	}
}

func (l *CKSNodeMetadataProtectionLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSNodeMetadataProtectionLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: metadata-protect
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: metadata-tester
  namespace: metadata-protect
spec:
  containers:
  - name: tester
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSNodeMetadataProtectionLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "metadata-protect", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to list network policies: %w", err)
	}
	if output == "" {
		return fmt.Errorf("no network policy found to protect node metadata")
	}
	return nil
}

func (l *CKSNodeMetadataProtectionLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create NetworkPolicy to block metadata access", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: block-metadata
  namespace: metadata-protect
spec:
  podSelector: {}
  policyTypes:
  - Egress
  egress:
  - to:
    - ipBlock:
        cidr: 0.0.0.0/0
        except:
        - 169.254.0.0/16
EOF`},
		{Description: "Verify policy exists", Command: "kubectl get networkpolicies -n metadata-protect"},
		{Description: "Test metadata access should fail", Command: "kubectl exec -n metadata-protect metadata-tester -- wget -q -O- --timeout=3 http://169.254.169.254/latest/meta-data/"},
	}
}
