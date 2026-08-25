package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&NetworkPolicyEgressDNSLab{}) }

type NetworkPolicyEgressDNSLab struct{ BaseLab }

func (l *NetworkPolicyEgressDNSLab) ID() string          { return "networkpolicy_egress_dns_blocked" }
func (l *NetworkPolicyEgressDNSLab) Title() string        { return "NetworkPolicy Blocks DNS Resolution" }
func (l *NetworkPolicyEgressDNSLab) Category() Category   { return CategoryNetworking }
func (l *NetworkPolicyEgressDNSLab) Difficulty() Difficulty { return DifficultyHard }
func (l *NetworkPolicyEgressDNSLab) EstimatedTime() int   { return 20 }
func (l *NetworkPolicyEgressDNSLab) Tags() []string {
	return []string{"networkpolicy", "dns", "egress", "networking"}
}
func (l *NetworkPolicyEgressDNSLab) Description() string {
	return `A deployment 'api' in namespace 'isolated' has a NetworkPolicy that
restricts egress to only external IPs (10.0.0.0/8). However, the app
needs DNS resolution (port 53 UDP/TCP to kube-dns in kube-system) to
function, and DNS queries are failing.

Your task: Add an egress rule to the NetworkPolicy that allows DNS
traffic (port 53 UDP+TCP) to pods in kube-system namespace, without
removing the existing external-only restriction.`
}
func (l *NetworkPolicyEgressDNSLab) Hints() []string {
	return []string{
		"Check: kubectl get netpol -n isolated -o yaml",
		"The existing egress rule allows 10.0.0.0/8 but NOT port 53",
		"Add a new egress rule with port 53 and port 53/udp",
		"Use toNamespaceSelector to match kube-system",
	}
}

func (l *NetworkPolicyEgressDNSLab) Break(ctx context.Context, kp string) error {
	if _, err := kubectl(ctx, kp, "create", "ns", "isolated"); err != nil {
		return err
	}
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: isolated
  labels:
    env: isolated
`
	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
  namespace: isolated
spec:
  replicas: 1
  selector:
    matchLabels:
      app: api
  template:
    metadata:
      labels:
        app: api
    spec:
      containers:
      - name: curl
        image: curlimages/curl:8.5.0
        command: ["sh","-c","while true; do nslookup kubernetes.default.svc.cluster.local 2>&1 || echo DNS_FAIL; sleep 5; done"]
`
	netpol := `apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: restrict-egress
  namespace: isolated
spec:
  podSelector:
    matchLabels:
      app: api
  policyTypes:
  - Egress
  egress:
  - to:
    - ipBlock:
        cidr: 10.0.0.0/8
`
	kubectlApply(ctx, kp, ns)
	kubectlApply(ctx, kp, deploy)
	return kubectlApply(ctx, kp, netpol)
}

func (l *NetworkPolicyEgressDNSLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(8 * time.Second)
	return nil
}

func (l *NetworkPolicyEgressDNSLab) Verify(ctx context.Context, kp string) error {
	logs, _ := kubectl(ctx, kp, "logs", "-l", "app=api", "-n", "isolated", "--tail=5")
	if strings.Contains(logs, "DNS_FAIL") {
		return fmt.Errorf("DNS still failing — logs show DNS_FAIL")
	}
	if strings.Contains(logs, "dns lookup") || strings.Contains(logs, "Address") {
		return nil // DNS working
	}
	if logs == "" {
		return fmt.Errorf("no logs yet — pod may not be ready")
	}
	return nil
}

func (l *NetworkPolicyEgressDNSLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Verify DNS failure", Command: "kubectl logs -l app=api -n isolated --tail=5", Notes: "Shows DNS_FAIL lines"},
		{Description: "Inspect the NetworkPolicy", Command: "kubectl get netpol restrict-egress -n isolated -o yaml", Notes: "Only has ipBlock 10.0.0.0/8, no port 53 rule"},
		{Description: "Patch to add DNS egress rule", Command: `kubectl patch netpol restrict-egress -n isolated --type=json -p='[{"op":"add","path":"/spec/egress/1","value":{"ports":[{"protocol":"UDP","port":53},{"protocol":"TCP","port":53}],"to":[{"namespaceSelector":{"matchLabels":{"kubernetes.io/metadata.name":"kube-system"}}}}]}]'`, Notes: "Adds UDP+TCP port 53 to kube-system"},
		{Description: "Verify DNS works", Command: "kubectl logs -l app=api -n isolated --tail=3", Notes: "No more DNS_FAIL lines"},
	}
}
