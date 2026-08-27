package labs

import (
	"context"
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
	return `Container images in the registry are not signed using Notary. This means
there is no way to verify the authenticity and integrity of images.

Your task: Configure Notary for image signing and sign the 'nginx:1.19'
image in the private registry.`
}

func (l *CKSImageNotarySignLab) Hints() []string {
	return []string{
		"Install Notary server and signer",
		"Initialize a Notary collection for the image",
		"Use notary-signer to sign the image",
	}
}

func (l *CKSImageNotarySignLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSImageNotarySignLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSImageNotarySignLab) Verify(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSImageNotarySignLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Install Notary", Command: "docker pull registry:2 notary/notaryserver notary/notarysigner"},
		{Description: "Initialize Notary", Command: "notary -s https://notary.example.com -d ~/.docker/trust init ./targets"},
		{Description: "Add target", Command: "notary -s https://notary.example.com -d ~/.docker/trust add targets/nginx ./nginx.tar.gz"},
		{Description: "Sign image", Command: "notary -s https://notary.example.com -d ~/.docker/trust publish targets/nginx"},
	}
}
