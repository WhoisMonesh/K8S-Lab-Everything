package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&NetworkPolicyAuditMode{})
}

type NetworkPolicyAuditMode struct {
	BaseLab
}

func (l *NetworkPolicyAuditMode) ID() string             { return "network_policy_audit_mode" }
func (l *NetworkPolicyAuditMode) Title() string          { return "NetworkPolicy in Audit Mode" }
func (l *NetworkPolicyAuditMode) Category() Category     { return CategoryNetworking }
func (l *NetworkPolicyAuditMode) Difficulty() Difficulty { return DifficultyMedium }
func (l *NetworkPolicyAuditMode) EstimatedTime() int     { return 20 }
func (l *NetworkPolicyAuditMode) Tags() []string         { return []string{"networking", "policy", "audit"} }

func (l *NetworkPolicyAuditMode) Description() string {
	return `A NetworkPolicy is configured in audit mode but traffic is still being blocked.
Debug and fix the NetworkPolicy to allow expected traffic.`
}

func (l *NetworkPolicyAuditMode) Hints() []string {
	return []string{
		"Check the NetworkPolicy policyTypes",
		"Verify ingress/egress rules",
		"Check if the policy has both Ingress and Egress in policyTypes",
	}
}

func (l *NetworkPolicyAuditMode) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *NetworkPolicyAuditMode) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: audit-test
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
  namespace: audit-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: web
  namespace: audit-test
spec:
  selector:
    app: web
  ports:
  - port: 80
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
  namespace: audit-test
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *NetworkPolicyAuditMode) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicy", "deny-all", "-n", "audit-test",
		"-o", "jsonpath={.spec.ingress}")
	if err != nil {
		return err
	}
	if output == "[]" || output == "" {
		return fmt.Errorf("network policy has no ingress rules")
	}
	return nil
}

func (l *NetworkPolicyAuditMode) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check NetworkPolicy", Command: "kubectl get networkpolicy deny-all -n audit-test -o yaml"},
		{Description: "Add ingress rule", Command: "kubectl patch networkpolicy deny-all -n audit-test --type='json' -p='[{\"op\": \"add\", \"path\": \"/spec/ingress/-\", \"value\": {\"from\": [{\"podSelector\": {}}]}}]'"},
	}
}
