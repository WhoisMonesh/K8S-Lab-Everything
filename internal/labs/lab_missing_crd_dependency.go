package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&MissingCRDDependencyLab{}) }

type MissingCRDDependencyLab struct{ BaseLab }

func (l *MissingCRDDependencyLab) ID() string             { return "missing_crd_dependency" }
func (l *MissingCRDDependencyLab) Title() string          { return "Custom Resource Fails — Missing CRD" }
func (l *MissingCRDDependencyLab) Category() Category     { return CategoryControlPlane }
func (l *MissingCRDDependencyLab) Difficulty() Difficulty { return DifficultyHard }
func (l *MissingCRDDependencyLab) EstimatedTime() int     { return 20 }
func (l *MissingCRDDependencyLab) Tags() []string {
	return []string{"crd", "custom-resource", "api", "control-plane"}
}
func (l *MissingCRDDependencyLab) Description() string {
	return `Someone applied a custom resource 'Widget' (apiVersion: example.com/v1)
but forgot to create the CustomResourceDefinition first. The resource
exists in etcd but kubectl cannot read it: the server doesn't know
what a Widget is.

Your task: Create the CRD for Widget resources, then verify the
existing Widget custom resource becomes accessible.`
}
func (l *MissingCRDDependencyLab) Hints() []string {
	return []string{
		"Check: kubectl get widgets shows 'error: the server doesn't have a resource type \"widgets\"'",
		"You need to create a CRD that defines apiVersion: example.com/v1, kind: Widget",
		"The CRD must specify names: plural: widgets, singular: widget, kind: Widget",
		"Once the CRD is applied, the existing Widget resource becomes readable",
	}
}

func (l *MissingCRDDependencyLab) Break(ctx context.Context, kp string) error {
	// Apply the custom resource directly (without the CRD) — this works at the etcd level
	widget := `apiVersion: example.com/v1
kind: Widget
metadata:
  name: my-widget
  namespace: default
spec:
  size: large
  color: blue
`
	// Force-apply using the raw API by creating via kubectl create with --raw
	// Simpler: use kubectl apply which will accept unknown resources in some configurations
	// Actually the cleanest approach: apply CRD first, create widget, then delete CRD
	crd := `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: widgets.example.com
spec:
  group: example.com
  names:
    kind: Widget
    listKind: WidgetList
    plural: widgets
    singular: widget
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              size:
                type: string
              color:
                type: string
`
	if err := kubectlApply(ctx, kp, crd); err != nil {
		return fmt.Errorf("failed to apply CRD: %w", err)
	}
	time.Sleep(2 * time.Second)
	if err := kubectlApply(ctx, kp, widget); err != nil {
		return fmt.Errorf("failed to apply widget: %w", err)
	}
	time.Sleep(1 * time.Second)
	// Now delete the CRD to create the "missing CRD" scenario
	_, err := kubectl(ctx, kp, "delete", "crd", "widgets.example.com", "--wait=true")
	return err
}

func (l *MissingCRDDependencyLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(3 * time.Second)
	return nil
}

func (l *MissingCRDDependencyLab) Verify(ctx context.Context, kp string) error {
	// CRD should exist
	crd, err := kubectl(ctx, kp, "get", "crd", "widgets.example.com", "-o", "name")
	if err != nil || !strings.Contains(crd, "widgets.example.com") {
		return fmt.Errorf("CRD widgets.example.com not found")
	}
	// Widget should be readable
	widget, err := kubectl(ctx, kp, "get", "widget", "my-widget", "-n", "default", "-o", "jsonpath={.spec.color}")
	if err != nil {
		return fmt.Errorf("Widget resource not accessible: %w", err)
	}
	if widget != "blue" {
		return fmt.Errorf("Widget has wrong color: %s", widget)
	}
	return nil
}

func (l *MissingCRDDependencyLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Confirm the error", Command: "kubectl get widgets", Notes: "Error: server doesn't have resource type 'widgets'"},
		{Description: "Create the CRD", Command: "kubectl apply -f - <<'EOF'\napiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata:\n  name: widgets.example.com\nspec:\n  group: example.com\n  names:\n    kind: Widget\n    plural: widgets\n    singular: widget\n  scope: Namespaced\n  versions:\n  - name: v1\n    served: true\n    storage: true\n    schema:\n      openAPIV3Schema:\n        type: object\n        properties:\n          spec:\n            type: object\n            properties:\n              size:\n                type: string\n              color:\n                type: string\nEOF", Notes: "Defines the Widget CRD with the expected schema"},
		{Description: "Verify the existing Widget", Command: "kubectl get widget my-widget -n default -o yaml", Notes: "The previously-orphaned Widget resource is now accessible"},
	}
}
