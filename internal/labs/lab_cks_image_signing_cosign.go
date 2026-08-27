package labs

import (
	"context"
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
	return `Container images are deployed without cryptographic signing. This allows
unsigned or tampered images to be deployed in the cluster.

Your task: Generate a cosign key pair and sign the 'nginx:1.19' image
with cosign to establish image integrity.`
}

func (l *CKSImageSigningCosignLab) Hints() []string {
	return []string{
		"Install cosign from sigstore",
		"Generate a key pair with cosign generate-key-pair",
		"Sign the image with cosign sign",
	}
}

func (l *CKSImageSigningCosignLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSImageSigningCosignLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSImageSigningCosignLab) Verify(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSImageSigningCosignLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Install cosign", Command: "go install github.com/sigstore/cosign/v2/cmd/cosign@latest"},
		{Description: "Generate key pair", Command: "cosign generate-key-pair"},
		{Description: "Sign the image", Command: "cosign sign --key cosign.key registry.example.com/nginx:1.19"},
		{Description: "Verify signature", Command: "cosign verify --key cosign.pub registry.example.com/nginx:1.19"},
	}
}
