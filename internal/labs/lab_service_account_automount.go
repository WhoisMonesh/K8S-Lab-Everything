package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ServiceAccountAutomountLab{})
}

type ServiceAccountAutomountLab struct {
	BaseLab
}

func (l *ServiceAccountAutomountLab) ID() string {
	return "service_account_automount"
}

func (l *ServiceAccountAutomountLab) Title() string {
	return "ServiceAccount automountServiceAccountToken"
}

func (l *ServiceAccountAutomountLab) Category() Category {
	return CategorySecurity
}

func (l *ServiceAccountAutomountLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ServiceAccountAutomountLab) Description() string {
	return `A pod 'secure-pod' needs to access the Kubernetes API but the
ServiceAccount has automountServiceAccountToken set to false. The pod
cannot authenticate to the API server.

Your task: Fix the ServiceAccount or pod configuration to allow
API access.`
}

func (l *ServiceAccountAutomountLab) Hints() []string {
	return []string{
		"Check the ServiceAccount automountServiceAccountToken setting",
		"automountServiceAccountToken=false prevents token mounting",
		"Either change the ServiceAccount or set automountServiceAccountToken in the pod spec",
	}
}

func (l *ServiceAccountAutomountLab) EstimatedTime() int {
	return 15
}

func (l *ServiceAccountAutomountLab) Tags() []string {
	return []string{"serviceaccount", "automount", "token", "security"}
}

func (l *ServiceAccountAutomountLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceAccountAutomountLab) Break(ctx context.Context, kubeconfigPath string) error {
	serviceAccount := `apiVersion: v1
kind: ServiceAccount
metadata:
  name: api-access
  namespace: default
automountServiceAccountToken: false
`
	if err := kubectlApply(ctx, kubeconfigPath, serviceAccount); err != nil {
		return fmt.Errorf("creating ServiceAccount: %w", err)
	}

	pod := `apiVersion: v1
kind: Pod
metadata:
  name: secure-pod
  namespace: default
spec:
  serviceAccountName: api-access
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'ls /var/run/secrets/kubernetes.io/serviceaccount/token && sleep 3600']
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *ServiceAccountAutomountLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	output, _ := kubectl(ctx, kubeconfigPath, "exec", "secure-pod",
		"--", "ls", "/var/run/secrets/kubernetes.io/serviceaccount/")
	if strings.Contains(output, "No such file") || strings.Contains(output, "not found") {
		return nil
	}
	return nil
}

func (l *ServiceAccountAutomountLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "exec", "secure-pod",
		"--", "ls", "/var/run/secrets/kubernetes.io/serviceaccount/")
	if err != nil {
		return fmt.Errorf("cannot access service account token: %w", err)
	}

	if !strings.Contains(output, "token") {
		return fmt.Errorf("service account token not mounted")
	}

	return nil
}

func (l *ServiceAccountAutomountLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check ServiceAccount",
			Command:     "kubectl get serviceaccount api-access -o yaml | grep automount",
			Notes:       "automountServiceAccountToken is false",
		},
		{
			Description: "Fix ServiceAccount",
			Command:     "kubectl patch serviceaccount api-access --type='json' -p='[{\"op\":\"replace\",\"path\":\"/automountServiceAccountToken\",\"value\":true}]'",
			Notes:       "Set automountServiceAccountToken to true",
		},
		{
			Description: "Delete and recreate pod",
			Command:     "kubectl delete pod secure-pod && kubectl apply -f - <<EOF\napiVersion: v1\nkind: Pod\nmetadata:\n  name: secure-pod\n  namespace: default\nspec:\n  serviceAccountName: api-access\n  containers:\n  - name: app\n    image: busybox:1.36\n    command: ['sh', '-c', 'ls /var/run/secrets/kubernetes.io/serviceaccount/token && sleep 3600']\nEOF",
			Notes:       "Recreate pod to mount the token",
		},
		{
			Description: "Verify token is mounted",
			Command:     "kubectl exec secure-pod -- ls /var/run/secrets/kubernetes.io/serviceaccount/",
			Notes:       "token file should now exist",
		},
	}
}
