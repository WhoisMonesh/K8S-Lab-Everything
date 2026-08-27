package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSImageMinimalBaseLab{})
}

type CKSImageMinimalBaseLab struct {
	BaseLab
}

func (l *CKSImageMinimalBaseLab) ID() string             { return "cks_image_minimal_base" }
func (l *CKSImageMinimalBaseLab) Title() string          { return "Use Minimal Base Images" }
func (l *CKSImageMinimalBaseLab) Category() Category     { return CategorySupplyChain }
func (l *CKSImageMinimalBaseLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *CKSImageMinimalBaseLab) EstimatedTime() int     { return 15 }
func (l *CKSImageMinimalBaseLab) Cert() Cert             { return CertCKS }
func (l *CKSImageMinimalBaseLab) DomainWeight() int      { return 20 }
func (l *CKSImageMinimalBaseLab) Tags() []string {
	return []string{"cks", "minimal-image", "distroless", "supply-chain"}
}

func (l *CKSImageMinimalBaseLab) Description() string {
	return `A deployment 'webapp' in namespace 'supply-chain-ns' uses the full 'ubuntu:20.04'
image. Full OS images include many unnecessary packages that increase the attack surface.

Your task: Update the deployment to use a minimal distroless image
'gcr.io/distroless/static:nonroot' instead.`
}

func (l *CKSImageMinimalBaseLab) Hints() []string {
	return []string{
		"Use kubectl set image to update the container image",
		"Distroless images have no shell, adjust command accordingly",
		"Verify the deployment rolls out successfully",
	}
}

func (l *CKSImageMinimalBaseLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSImageMinimalBaseLab) Break(ctx context.Context, kubeconfigPath string) error {
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
  name: webapp
  namespace: supply-chain-ns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: web
        image: ubuntu:20.04
        command: ["sleep", "3600"]
`
	return kubectlApply(ctx, kubeconfigPath, deploy)
}

func (l *CKSImageMinimalBaseLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "webapp", "-n", "supply-chain-ns",
		"-o", "jsonpath={.spec.template.spec.containers[0].image}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if strings.Contains(output, "distroless") {
		return nil
	}
	return fmt.Errorf("deployment not using minimal image (got: %s)", output)
}

func (l *CKSImageMinimalBaseLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Update image to distroless", Command: "kubectl set image deployment/webapp web=gcr.io/distroless/static:nonroot -n supply-chain-ns"},
		{Description: "Verify rollout", Command: "kubectl rollout status deployment/webapp -n supply-chain-ns"},
		{Description: "Check image", Command: "kubectl get deployment webapp -n supply-chain-ns -o jsonpath='{.spec.template.spec.containers[0].image}'"},
	}
}
