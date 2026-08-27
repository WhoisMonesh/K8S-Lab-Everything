package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSImagePullPolicyAlwaysLab{})
}

type CKSImagePullPolicyAlwaysLab struct {
	BaseLab
}

func (l *CKSImagePullPolicyAlwaysLab) ID() string             { return "cks_image_pull_policy_always" }
func (l *CKSImagePullPolicyAlwaysLab) Title() string          { return "Enforce Always Pull Policy" }
func (l *CKSImagePullPolicyAlwaysLab) Category() Category     { return CategorySupplyChain }
func (l *CKSImagePullPolicyAlwaysLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKSImagePullPolicyAlwaysLab) EstimatedTime() int     { return 15 }
func (l *CKSImagePullPolicyAlwaysLab) Cert() Cert             { return CertCKS }
func (l *CKSImagePullPolicyAlwaysLab) DomainWeight() int      { return 20 }
func (l *CKSImagePullPolicyAlwaysLab) Tags() []string {
	return []string{"cks", "image-pull", "supply-chain", "security"}
}

func (l *CKSImagePullPolicyAlwaysLab) Description() string {
	return `Pods in the 'supply-chain-ns' namespace use imagePullPolicy: IfNotPresent.
This means cached images are used without verification, potentially allowing
malicious images to run.

Your task: Patch the deployments in the namespace to use imagePullPolicy: Always
for all containers.`
}

func (l *CKSImagePullPolicyAlwaysLab) Hints() []string {
	return []string{
		"Use kubectl patch to update deployment specs",
		"Set imagePullPolicy to Always for each container",
		"Verify the patch was applied correctly",
	}
}

func (l *CKSImagePullPolicyAlwaysLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSImagePullPolicyAlwaysLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: supply-chain-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-app
  namespace: supply-chain-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web-app
  template:
    metadata:
      labels:
        app: web-app
    spec:
      containers:
      - name: web
        image: nginx:1.19
        imagePullPolicy: IfNotPresent
`
	return kubectlApply(ctx, kubeconfigPath, deploy)
}

func (l *CKSImagePullPolicyAlwaysLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "web-app", "-n", "supply-chain-ns",
		"-o", "jsonpath={.spec.template.spec.containers[0].imagePullPolicy}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if strings.TrimSpace(output) == "Always" {
		return nil
	}
	return fmt.Errorf("imagePullPolicy not set to Always (got: %s)", output)
}

func (l *CKSImagePullPolicyAlwaysLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Patch deployment to Always pull", Command: "kubectl patch deployment web-app -n supply-chain-ns --type='json' -p='[{\"op\": \"replace\", \"path\": \"/spec/template/spec/containers/0/imagePullPolicy\", \"value\": \"Always\"}]'"},
		{Description: "Verify", Command: "kubectl get deployment web-app -n supply-chain-ns -o jsonpath='{.spec.template.spec.containers[0].imagePullPolicy}'"},
	}
}
