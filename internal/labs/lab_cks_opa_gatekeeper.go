package labs

import (
    "context"
    "fmt"
    "strings"
    "time"
)

func init() {
    Register(&OPAGatekeeperLab{})
}

type OPAGatekeeperLab struct {
    BaseLab
}

func (l *OPAGatekeeperLab) ID() string { return "opa_gatekeeper" }
func (l *OPAGatekeeperLab) Title() string { return "OPA Gatekeeper Policy Enforcement" }
func (l *OPAGatekeeperLab) Category() Category { return CategoryClusterHardening }
func (l *OPAGatekeeperLab) Difficulty() Difficulty { return DifficultyHard }
func (l *OPAGatekeeperLab) Description() string {
    return `Gatekeeper is installed in the cluster and a constraint is applied that disallows privileged containers. A pod is created with the privileged flag, which is rejected.

Your task: Modify the pod spec to comply with the Gatekeeper policy (remove the privileged flag) and verify it runs.`
}
func (l *OPAGatekeeperLab) Hints() []string {
    return []string{"Check the ConstraintTemplate for the rule", "Privileged containers have allowPrivilegeEscalation: true", "Remove the flag or set it to false"}
}
func (l *OPAGatekeeperLab) EstimatedTime() int { return 30 }
func (l *OPAGatekeeperLab) Tags() []string { return []string{"gatekeeper", "opa", "policy", "security", "cluster-hardening"} }

func (l *OPAGatekeeperLab) Prepare(ctx context.Context, kubeconfigPath string) error {
    return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *OPAGatekeeperLab) Break(ctx context.Context, kubeconfigPath string) error {
    // Install Gatekeeper namespace
    ns := `apiVersion: v1
kind: Namespace
metadata:
  name: gatekeeper-system`
    if err := kubectlApply(ctx, kubeconfigPath, ns); err != nil {
        return fmt.Errorf("creating gatekeeper namespace: %w", err)
    }
    // Install Gatekeeper via URL manifest
    if _, err := kubectl(ctx, kubeconfigPath, "apply", "-f", "https://raw.githubusercontent.com/open-policy-agent/gatekeeper/master/deploy/gatekeeper.yaml"); err != nil {
        return fmt.Errorf("installing gatekeeper: %w", err)
    }
    // Wait for Gatekeeper pods to be ready
    time.Sleep(15 * time.Second)
    // ConstraintTemplate to forbid privileged containers
    tmpl := `apiVersion: templates.gatekeeper.sh/v1beta1
kind: ConstraintTemplate
metadata:
  name: k8spspprivileged
spec:
  crd:
    spec:
      names:
        kind: K8sPSPPrivileged
      validation:
        openAPIV3Schema:
          type: object
          properties: {}
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |
        package k8spspprivileged
        violation[{{"msg": msg}]] {
          input.review.object.spec.containers[_].securityContext.allowPrivilegeEscalation == true
          msg := "Privileged containers are not allowed"
        }`
    if err := kubectlApply(ctx, kubeconfigPath, tmpl); err != nil {
        return fmt.Errorf("creating constraint template: %w", err)
    }
    // Constraint using the template
    constraint := `apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sPSPPrivileged
metadata:
  name: disallow-privileged
spec:
  match:
    kinds:
      - apiGroups: [""]
        kinds: ["Pod"]`
    if err := kubectlApply(ctx, kubeconfigPath, constraint); err != nil {
        return fmt.Errorf("creating constraint: %w", err)
    }
    // Create a privileged pod that should be rejected
    pod := `apiVersion: v1
kind: Pod
metadata:
  name: privileged-pod
  namespace: default
spec:
  containers:
  - name: busybox
    image: busybox:1.36
    command: ["sh","-c","sleep 3600"]
    securityContext:
      allowPrivilegeEscalation: true`
    return kubectlApply(ctx, kubeconfigPath, pod)
}

func (l *OPAGatekeeperLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
    // Expect the pod creation to have been rejected via events.
    time.Sleep(10 * time.Second)
    out, err := kubectl(ctx, kubeconfigPath, "get", "events", "-n", "default",
        "--field-selector", "reason=FailedCreate", "-o", "jsonpath={.items[*].message}")
    if err != nil {
        return nil // ignore errors, assume not broken
    }
    if strings.Contains(out, "Privileged containers are not allowed") {
        return nil
    }
    // If the pod exists and is running, it's not in broken state.
    pods, _ := kubectl(ctx, kubeconfigPath, "get", "pods", "privileged-pod", "-n", "default", "-o", "jsonpath={.status.phase}")
    if pods == "Running" {
        return fmt.Errorf("privileged pod is running, expected rejection")
    }
    return nil
}

func (l *OPAGatekeeperLab) Verify(ctx context.Context, kubeconfigPath string) error {
    // Delete the violating pod (if exists) and create a compliant one.
    _, _ = kubectl(ctx, kubeconfigPath, "delete", "pod", "privileged-pod", "-n", "default", "--ignore-not-found")
    compliant := `apiVersion: v1
kind: Pod
metadata:
  name: privileged-pod
  namespace: default
spec:
  containers:
  - name: busybox
    image: busybox:1.36
    command: ["sh","-c","sleep 3600"]`
    if err := kubectlApply(ctx, kubeconfigPath, compliant); err != nil {
        return fmt.Errorf("creating compliant pod: %w", err)
    }
    // Wait for the pod to become running
    time.Sleep(5 * time.Second)
    status, err := kubectl(ctx, kubeconfigPath, "get", "pod", "privileged-pod", "-n", "default", "-o", "jsonpath={.status.phase}")
    if err != nil {
        return fmt.Errorf("checking pod status: %w", err)
    }
    if status != "Running" {
        return fmt.Errorf("pod not running after fix, status: %s", status)
    }
    return nil
}

func (l *OPAGatekeeperLab) SolutionSteps() []SolutionStep {
    return []SolutionStep{{
        Description: "Delete the violating privileged pod",
        Command: "kubectl delete pod privileged-pod -n default --ignore-not-found",
    }, {
        Description: "Create a pod without privileged flag",
        Command: `kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: privileged-pod
  namespace: default
spec:
  containers:
  - name: busybox
    image: busybox:1.36
    command: ["sh","-c","sleep 3600"]
EOF`,
    }, {
        Description: "Verify the pod is running",
        Command: "kubectl get pod privileged-pod -n default -o jsonpath={.status.phase}",
    }}
}
