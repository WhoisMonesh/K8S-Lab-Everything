package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
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
	return `An SBOM record for the application image must be stored in the cluster so that
component dependencies can be tracked for vulnerability management. The ConfigMap
'sbom' in namespace 'sbom-ns' currently has an empty 'sbom' field.

Your task: Generate a minimal SBOM for the image and store the result in the
'sbom' data field of the ConfigMap so it is non-empty.`
}

func (l *CKSSBOMGenerationLab) Hints() []string {
	return []string{
		"Generate a minimal SBOM for the image",
		"Store the SBOM text in the 'sbom' data field of the ConfigMap",
		"Verify the field is no longer empty",
	}
}

func (l *CKSSBOMGenerationLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSSBOMGenerationLab) Break(ctx context.Context, kubeconfigPath string) error {
	ns := `apiVersion: v1
kind: Namespace
metadata:
  name: sbom-ns
`
	if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	cm := `apiVersion: v1
kind: ConfigMap
metadata:
  name: sbom
  namespace: sbom-ns
data:
  image: "nginx"
  sbom: ""
`
	return kubectlApply(ctx, kubeconfigPath, cm)
}

func (l *CKSSBOMGenerationLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *CKSSBOMGenerationLab) Verify(ctx context.Context, kubeconfigPath string) error {
	sbom, err := kubectl(ctx, kubeconfigPath, "get", "configmap", "sbom", "-n", "sbom-ns",
		"-o", "jsonpath={.data.sbom}")
	if err != nil {
		return fmt.Errorf("failed to get configmap: %w", err)
	}
	if strings.TrimSpace(sbom) == "" {
		return fmt.Errorf("SBOM 'sbom' field is empty; no SBOM stored")
	}
	return nil
}

func (l *CKSSBOMGenerationLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create/update the configmap with a minimal SBOM", Command: `kubectl create configmap sbom -n sbom-ns --from-literal=image=nginx --from-literal='sbom=SPDXVersion: SPDX-2.3
DataLicense: CC0-1.0
SPDXID: SPDXRef-DOCUMENT
DocumentName: nginx-sbom
Packages: nginx:1.19' --dry-run=client -o yaml | kubectl apply -f -`},
		{Description: "Verify the SBOM field is non-empty", Command: "kubectl get configmap sbom -n sbom-ns -o jsonpath='{.data.sbom}'"},
	}
}
