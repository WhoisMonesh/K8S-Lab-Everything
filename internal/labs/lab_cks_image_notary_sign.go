package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSImageNotarySignLab{})
}

type CKSImageNotarySignLab struct {
	BaseLab
}

func (l *CKSImageNotarySignLab) ID() string             { return "cks_image_notary_sign" }
func (l *CKSImageNotarySignLab) Title() string          { return "Sign Images with Notary" }
func (l *CKSImageNotarySignLab) Category() Category     { return CategorySupplyChain }
func (l *CKSImageNotarySignLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSImageNotarySignLab) EstimatedTime() int     { return 30 }
func (l *CKSImageNotarySignLab) Cert() Cert             { return CertCKS }
func (l *CKSImageNotarySignLab) DomainWeight() int      { return 20 }
func (l *CKSImageNotarySignLab) Tags() []string {
	return []string{"cks", "notary", "image-signing", "supply-chain"}
}

func (l *CKSImageNotarySignLab) Description() string {
	return `A deployment 'app' in namespace 'trusted-app' references an image from the
private registry 'registry.example.com' but has no imagePullSecrets configured,
so it stays in ImagePullBackOff and cannot be trusted/pulled securely.

Your task: Create a docker registry secret named 'regcred' in namespace
'trusted-app' and attach it to the deployment via imagePullSecrets so the
deployment can pull the private image.`
}

func (l *CKSImageNotarySignLab) Hints() []string {
	return []string{
		"Create a docker-registry secret named regcred",
		"Add the secret to the deployment's imagePullSecrets",
		"Patch the deployment to reference the secret",
	}
}

func (l *CKSImageNotarySignLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSImageNotarySignLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: trusted-app
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  namespace: trusted-app
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
        image: registry.example.com/nginx:1.19
`
	return kubectlApply(ctx, kubeconfigPath, deploy)
}

func (l *CKSImageNotarySignLab) VerifyBroken(_ context.Context, _ string) error {
	return nil
}

func (l *CKSImageNotarySignLab) Verify(ctx context.Context, kubeconfigPath string) error {
	secret, err := kubectl(ctx, kubeconfigPath, "get", "secret", "regcred", "-n", "trusted-app", "-o", "name")
	if err != nil {
		return fmt.Errorf("secret 'regcred' not found in namespace 'trusted-app'")
	}
	if strings.TrimSpace(secret) == "" {
		return fmt.Errorf("secret 'regcred' not found in namespace 'trusted-app'")
	}
	imgPull, err := kubectl(ctx, kubeconfigPath, "get", "deployment", "app", "-n", "trusted-app",
		"-o", "jsonpath={.spec.template.spec.imagePullSecrets[*].name}")
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}
	if !strings.Contains(imgPull, "regcred") {
		return fmt.Errorf("deployment does not reference 'regcred' in imagePullSecrets")
	}
	return nil
}

func (l *CKSImageNotarySignLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create the docker registry secret", Command: "kubectl create secret docker-registry regcred -n trusted-app --docker-server=registry.example.com --docker-username=<user> --docker-password=<password> --docker-email=<email>"},
		{Description: "Attach the secret to the deployment", Command: "kubectl patch deployment app -n trusted-app -p '{\"spec\":{\"template\":{\"spec\":{\"imagePullSecrets\":[{\"name\":\"regcred\"}]}}}}'"},
		{Description: "Verify the secret is referenced", Command: "kubectl get deployment app -n trusted-app -o jsonpath='{.spec.template.spec.imagePullSecrets[*].name}'"},
	}
}
