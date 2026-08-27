package labs

import (
	"context"
)

func init() {
	Register(&CKSSBOMGenerationLab{})
}

type CKSSBOMGenerationLab struct {
	BaseLab
}

func (l *CKSSBOMGenerationLab) ID() string             { return "cks_sbom_generation" }
func (l *CKSSBOMGenerationLab) Title() string          { return "Generate SBOM for Container Images" }
func (l *CKSSBOMGenerationLab) Category() Category     { return CategorySupplyChain }
func (l *CKSSBOMGenerationLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSSBOMGenerationLab) EstimatedTime() int     { return 25 }
func (l *CKSSBOMGenerationLab) Cert() Cert             { return CertCKS }
func (l *CKSSBOMGenerationLab) DomainWeight() int      { return 20 }
func (l *CKSSBOMGenerationLab) Tags() []string {
	return []string{"cks", "sbom", "supply-chain", "spdx"}
}

func (l *CKSSBOMGenerationLab) Description() string {
	return `Container images do not have Software Bill of Materials (SBOM). Without SBOMs,
it is impossible to track dependencies and identify known vulnerabilities.

Your task: Use Syft to generate an SBOM for the 'nginx:1.19' image in SPDX format.`
}

func (l *CKSSBOMGenerationLab) Hints() []string {
	return []string{
		"Install syft from anchore",
		"Run syft with --output spdx-json format",
		"Save the SBOM to a file",
	}
}

func (l *CKSSBOMGenerationLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSSBOMGenerationLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSSBOMGenerationLab) Verify(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSSBOMGenerationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Install syft", Command: "curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin"},
		{Description: "Generate SBOM", Command: "syft nginx:1.19 -o spdx-json > nginx-sbom.spdx.json"},
		{Description: "View SBOM", Command: "cat nginx-sbom.spdx.json | jq '.packages | length'"},
		{Description: "Scan SBOM for vulnerabilities", Command: "grype sbom:nginx-sbom.spdx.json"},
	}
}
