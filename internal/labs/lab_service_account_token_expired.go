package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&ServiceAccountTokenExpired{})
}

type ServiceAccountTokenExpired struct {
	BaseLab
}

func (l *ServiceAccountTokenExpired) ID() string             { return "service_account_token_expired" }
func (l *ServiceAccountTokenExpired) Title() string          { return "Service Account Token Expired" }
func (l *ServiceAccountTokenExpired) Category() Category     { return CategorySecurity }
func (l *ServiceAccountTokenExpired) Difficulty() Difficulty { return DifficultyMedium }
func (l *ServiceAccountTokenExpired) EstimatedTime() int     { return 15 }
func (l *ServiceAccountTokenExpired) Tags() []string {
	return []string{"security", "serviceaccount", "tokens"}
}

func (l *ServiceAccountTokenExpired) Description() string {
	return `A pod is failing because it's using an expired service account token.
The pod needs to authenticate to the API server but the mounted token is invalid.`
}

func (l *ServiceAccountTokenExpired) Hints() []string {
	return []string{
		"Check the pod's service account",
		"Look at the mounted secrets",
		"Delete the old secret to trigger token regeneration",
	}
}

func (l *ServiceAccountTokenExpired) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceAccountTokenExpired) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: api-caller
---
apiVersion: v1
kind: Secret
metadata:
  name: api-caller-token-old
  annotations:
    kubernetes.io/service-account.name: api-caller
  labels:
    kubernetes.io/service-account.name: api-caller
type: kubernetes.io/service-account-token
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-client
spec:
  replicas: 1
  selector:
    matchLabels:
      app: api-client
  template:
    metadata:
      labels:
        app: api-client
    spec:
      serviceAccountName: api-caller
      containers:
      - name: curl
        image: curlimages/curl
        command: ["sleep", "3600"]
        volumeMounts:
        - name: token
          mountPath: /var/run/secrets/kubernetes.io/serviceaccount
          readOnly: true
      volumes:
      - name: token
        secret:
          secretName: api-caller-token-old
          items:
          - key: token
            path: token`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *ServiceAccountTokenExpired) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "deploy/api-client",
		"--", "cat", "/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return err
	}
	if len(output) < 50 {
		return fmt.Errorf("token appears invalid")
	}
	return nil
}

func (l *ServiceAccountTokenExpired) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check service account secrets", Command: "kubectl describe sa api-caller"},
		{Description: "Delete old secret", Command: "kubectl delete secret api-caller-token-old"},
		{Description: "Restart pod to get new token", Command: "kubectl rollout restart deployment api-client"},
	}
}
