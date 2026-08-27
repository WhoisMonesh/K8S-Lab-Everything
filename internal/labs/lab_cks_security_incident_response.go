package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSSecurityIncidentResponseLab{})
}

type CKSSecurityIncidentResponseLab struct {
	BaseLab
}

func (l *CKSSecurityIncidentResponseLab) ID() string             { return "cks_security_incident_response" }
func (l *CKSSecurityIncidentResponseLab) Title() string          { return "Respond to Security Incidents" }
func (l *CKSSecurityIncidentResponseLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSSecurityIncidentResponseLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSSecurityIncidentResponseLab) EstimatedTime() int     { return 30 }
func (l *CKSSecurityIncidentResponseLab) Cert() Cert             { return CertCKS }
func (l *CKSSecurityIncidentResponseLab) DomainWeight() int      { return 20 }
func (l *CKSSecurityIncidentResponseLab) Tags() []string {
	return []string{"cks", "incident-response", "forensics", "monitoring"}
}

func (l *CKSSecurityIncidentResponseLab) Description() string {
	return `A compromised pod 'compromised-app' in namespace 'incident-response' is
running suspicious processes. It needs to be isolated and forensics collected
without destroying evidence.

Your task: 
1. Create a NetworkPolicy that isolates the pod (deny all traffic)
2. Create an ephemeral debug container to investigate
3. Capture the process list and network connections as evidence`
}

func (l *CKSSecurityIncidentResponseLab) Hints() []string {
	return []string{
		"Create a NetworkPolicy targeting the compromised pod",
		"Use kubectl debug for ephemeral containers",
		"Capture evidence to a PersistentVolume",
	}
}

func (l *CKSSecurityIncidentResponseLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSSecurityIncidentResponseLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: incident-response
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: compromised-app
  namespace: incident-response
  labels:
    app: compromised
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ["sh", "-c", "while true; do sleep 3600; done"]
`
	return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *CKSSecurityIncidentResponseLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "networkpolicies", "-n", "incident-response", "-o", "yaml")
	if err != nil {
		return fmt.Errorf("failed to get network policies: %w", err)
	}
	if !strings.Contains(output, "isolate-compromised") {
		return fmt.Errorf("isolation policy not found")
	}
	return nil
}

func (l *CKSSecurityIncidentResponseLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Isolate compromised pod", Command: `cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: isolate-compromised
  namespace: incident-response
spec:
  podSelector:
    matchLabels:
      app: compromised
  policyTypes:
  - Ingress
  - Egress
EOF`},
		{Description: "Debug the pod", Command: "kubectl debug compromised-app -n incident-response -it --image=busybox:1.36 -- sh"},
		{Description: "Capture evidence", Command: "kubectl exec -n incident-response compromised-app -- sh -c 'ps aux > /tmp/evidence.txt; netstat -tlnp >> /tmp/evidence.txt'"},
		{Description: "Verify isolation", Command: "kubectl get networkpolicies -n incident-response"},
	}
}
