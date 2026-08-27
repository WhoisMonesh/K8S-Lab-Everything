package labs

import (
	"context"
	"fmt"
)

func init() {
	Register(&MissingCRDDependency{})
}

type MissingCRDDependency struct {
	BaseLab
}

func (l *MissingCRDDependency) ID() string             { return "missing_crd_dependency2" }
func (l *MissingCRDDependency) Title() string          { return "Custom Resource Fails - Missing CRD" }
func (l *MissingCRDDependency) Category() Category     { return CategoryControlPlane }
func (l *MissingCRDDependency) Difficulty() Difficulty { return DifficultyHard }
func (l *MissingCRDDependency) EstimatedTime() int     { return 20 }
func (l *MissingCRDDependency) Tags() []string         { return []string{"crd", "custom-resources", "api"} }

func (l *MissingCRDDependency) Description() string {
	return `A Custom Resource is failing because the required CRD doesn't exist.
Create the CRD to allow the Custom Resource to work.`
}

func (l *MissingCRDDependency) Hints() []string {
	return []string{
		"Check if the CRD exists",
		"Look at the Custom Resource definition",
		"Create the required CRD",
	}
}

func (l *MissingCRDDependency) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *MissingCRDDependency) Break(ctx context.Context, kubeconfigPath string) error {
	manifest := `apiVersion: stable.example.com/v1
kind: CronTab
metadata:
  name: my-cron
spec:
  cronSpec: "* * * * */5"
  image: my-app`
	return kubectlApply(ctx, kubeconfigPath, manifest)
}

func (l *MissingCRDDependency) Verify(ctx context.Context, kubeconfigPath string) error {
	_, err := kubectl(ctx, kubeconfigPath, "get", "crd", "crontabs.stable.example.com")
	if err != nil {
		return fmt.Errorf("CRD not found")
	}
	return nil
}

func (l *MissingCRDDependency) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check CRDs", Command: "kubectl get crds"},
		{Description: "Create CRD", Command: "kubectl apply -f - <<EOF\napiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata:\n  name: crontabs.stable.example.com\nspec:\n  group: stable.example.com\n  versions:\n    - name: v1\n      served: true\n      storage: true\n      schema:\n        openAPIV3Schema:\n          type: object\n          properties:\n            spec:\n              type: object\n              properties:\n                cronSpec:\n                  type: string\n                image:\n                  type: string\n  scope: Namespaced\n  names:\n    plural: crontabs\n    singular: crontab\n    kind: CronTab\nEOF"},
		{Description: "Create Custom Resource", Command: "kubectl apply -f my-cron.yaml"},
	}
}
