package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&PodNetworkConnectivity{})
}

type PodNetworkConnectivity struct {
	BaseLab
}

func (l *PodNetworkConnectivity) ID() string             { return "pod_network_connectivity" }
func (l *PodNetworkConnectivity) Title() string          { return "Pod-to-Pod Network Connectivity" }
func (l *PodNetworkConnectivity) Category() Category     { return CategoryNetworking }
func (l *PodNetworkConnectivity) Difficulty() Difficulty { return DifficultyHard }
func (l *PodNetworkConnectivity) EstimatedTime() int     { return 25 }
func (l *PodNetworkConnectivity) Tags() []string {
	return []string{"networking", "pods", "connectivity"}
}

func (l *PodNetworkConnectivity) Description() string {
	return `Pods in different namespaces cannot communicate with each other.
Debug the network connectivity issue and fix the NetworkPolicy that is blocking traffic.`
}

func (l *PodNetworkConnectivity) Hints() []string {
	return []string{
		"Check NetworkPolicies in all namespaces",
		"Test connectivity with kubectl exec",
		"Look for policies blocking ingress/egress",
	}
}

func (l *PodNetworkConnectivity) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodNetworkConnectivity) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: Namespace
metadata:
  name: app-a
---
apiVersion: v1
kind: Namespace
metadata:
  name: app-b
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: server
  namespace: app-a
spec:
  replicas: 1
  selector:
    matchLabels:
      app: server
  template:
    metadata:
      labels:
        app: server
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
  name: server
  namespace: app-a
spec:
  selector:
    app: server
  ports:
  - port: 80
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: client
  namespace: app-b
spec:
  replicas: 1
  selector:
    matchLabels:
      app: client
  template:
    metadata:
      labels:
        app: client
    spec:
      containers:
      - name: curl
        image: curlimages/curl
        command: ["sleep", "3600"]
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: block-cross-namespace
  namespace: app-a
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: app-a`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *PodNetworkConnectivity) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "-n", "app-b", "deploy/client",
		"--", "curl", "-s", "--max-time", "5", "server.app-a.svc.cluster.local")
	if err != nil {
		return fmt.Errorf("connectivity test failed: %w", err)
	}
	if output == "" {
		return fmt.Errorf("no response from server")
	}
	return nil
}

func (l *PodNetworkConnectivity) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check NetworkPolicies", Command: "kubectl get networkpolicies -A"},
		{Description: "Test connectivity", Command: "kubectl exec -n app-b deploy/client -- curl -s server.app-a.svc.cluster.local"},
		{Description: "Delete blocking policy", Command: "kubectl delete networkpolicy block-cross-namespace -n app-a"},
	}
}
