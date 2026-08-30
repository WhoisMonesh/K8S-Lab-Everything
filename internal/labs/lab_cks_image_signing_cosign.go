package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&CKSImageSigningCosignLab{})
}

type CKSImageSigningCosignLab struct {
	BaseLab
}

func (l *CKSImageSigningCosignLab) ID() string             { return "cks_image_signing_cosign" }
func (l *CKSImageSigningCosignLab) Title() string          { return "Sign Images with Cosign" }
func (l *CKSImageSigningCosignLab) Category() Category     { return CategorySupplyChain }
func (l *CKSImageSigningCosignLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSImageSigningCosignLab) EstimatedTime() int     { return 30 }
func (l *CKSImageSigningCosignLab) Cert() Cert             { return CertCKS }
func (l *CKSImageSigningCosignLab) DomainWeight() int      { return 20 }
func (l *CKSImageSigningCosignLab) Tags() []string {
	return []string{"cks", "cosign", "image-signing", "supply-chain"}
}

func (l *CKSImageSigningCosignLab) Description() string {
	return `A deployment 'app' in namespace 'pinned-app' references the mutable image tag
'nginx:latest'. Mutable tags can be silently retagged to a malicious image.

Your task: Pin the deployment to an immutable image digest (containing '@sha256:')
and ensure the imagePullPolicy allows the pinned digest to be used.`
}

func (l *CKSImageSigningCosignLab) Hints() []string {
	return []string{
		"Resolve nginx:latest to its digest",
		"Set the deployment image to the full digest reference",
		"Ensure imagePullPolicy is not Always",
	}
}

func (l *CKSImageSigningCosignLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSImageSigningCosignLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: pinned-app
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: pinned-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: app
  template:
    metadata:
      labels:
        app: app
    spec:
      containers:
      - name: app
        image: nginx:latest
        imagePullPolicy: Always
`
	return kubectlApply(ctx, kubeconfigPath, deploy)
}

func (l *CKSImageSigningCosignLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *CKSImageSigningCosignLab) Verify(ctx context.Context, kubeconfigPath string) error {
	image, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "app", "-n", "pinned-app",
		"-o", "jsonpath={.spec.template.spec.containers[0].image}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if !strings.Contains(strings.TrimSpace(image), "@sha256:") {
		return fmt.Errorf("image is not pinned to an immutable digest (got: %s)", image)
	}
	pullPolicy, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "app", "-n", "pinned-app",
		"-o", "jsonpath={.spec.template.spec.containers[0].imagePullPolicy}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	pp := strings.TrimSpace(pullPolicy)
	if pp == "IfNotPresent" || pp == "Never" {
		return nil
	}
	if pp == "Always" {
		return fmt.Errorf("imagePullPolicy must not be Always for a pinned digest")
	}
	return nil
}

func (l *CKSImageSigningCosignLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Resolve nginx:latest to its digest", Command: "docker pull nginx:latest && docker inspect nginx:latest --format='{{index .RepoDigests 0}}'"},
		{Description: "Pin the deployment to the digest", Command: "kubectl set image deployment/app app=nginx@sha256:<digest> -n pinned-app"},
		{Description: "Verify the pin and pull policy", Command: "kubectl get deployment app -n pinned-app -o jsonpath='{.spec.template.spec.containers[0].image} {.spec.template.spec.containers[0].imagePullPolicy}'"},
	}
}
