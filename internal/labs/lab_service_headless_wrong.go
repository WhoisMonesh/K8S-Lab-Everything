package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&ServiceHeadlessWrongLab{})
}

type ServiceHeadlessWrongLab struct {
	BaseLab
}

func (l *ServiceHeadlessWrongLab) ID() string {
	return "service_headless_wrong"
}

func (l *ServiceHeadlessWrongLab) Title() string {
	return "Headless Service Selector Issue"
}

func (l *ServiceHeadlessWrongLab) Category() Category {
	return CategoryNetworking
}

func (l *ServiceHeadlessWrongLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *ServiceHeadlessWrongLab) Description() string {
	return `A headless Service 'database-hl' is supposed to provide DNS entries
for StatefulSet pods but has wrong selectors. The Service selector doesn't
match the pod labels, so no endpoints are created.

Your task: Fix the headless Service selector to match the pod labels.`
}

func (l *ServiceHeadlessWrongLab) Hints() []string {
	return []string{
		"Check the Service selector",
		"Compare Service selector with pod labels",
		"Headless services must have correct selectors to create endpoints",
	}
}

func (l *ServiceHeadlessWrongLab) EstimatedTime() int {
	return 15
}

func (l *ServiceHeadlessWrongLab) Tags() []string {
	return []string{"service", "headless", "dns", "statefulset", "networking"}
}

func (l *ServiceHeadlessWrongLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *ServiceHeadlessWrongLab) Break(ctx context.Context, kubeconfigPath string) error {
	statefulset := `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: database
  namespace: default
spec:
  serviceName: database-hl
  replicas: 2
  selector:
    matchLabels:
      app: database
  template:
    metadata:
      labels:
        app: database
        role: db
    spec:
      containers:
      - name: db
        image: busybox:1.36
        command: ['sh', '-c', 'sleep 3600']
`
	if err := kubectlApply(ctx, kubeconfigPath, statefulset); err != nil {
		return fmt.Errorf("creating statefulset: %w", err)
	}

	service := `apiVersion: v1
kind: Service
metadata:
  name: database-hl
  namespace: default
spec:
  clusterIP: None
  selector:
    app: database
    role: wrong-label
  ports:
  - port: 3306
    targetPort: 3306
`
	if err := kubectlApply(ctx, kubeconfigPath, service); err != nil {
		return fmt.Errorf("creating service: %w", err)
	}

	return nil
}

func (l *ServiceHeadlessWrongLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(15 * time.Second)
	return nil
}

func (l *ServiceHeadlessWrongLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "endpoints", "database-hl",
		"-o", "jsonpath={.subsets[*].addresses[*].ip}")
	if err != nil {
		return fmt.Errorf("failed to check endpoints: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no endpoints found")
	}

	// Test DNS resolution
	output, err = kubectl(ctx, kubeconfigPath, "exec", "database-0",
		"--", "nslookup", "database-hl.default.svc.cluster.local")
	if err != nil {
		return fmt.Errorf("DNS resolution failed: %w", err)
	}

	if !strings.Contains(output, "Address") {
		return fmt.Errorf("DNS not resolving")
	}

	return nil
}

func (l *ServiceHeadlessWrongLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check Service selector",
			Command:     "kubectl get service database-hl -o yaml | grep -A 5 selector",
			Notes:       "Selector has role=wrong-label which doesn't match pods",
		},
		{
			Description: "Check pod labels",
			Command:     "kubectl get pods --show-labels | grep database",
			Notes:       "Pods have role=db, not role=wrong-label",
		},
		{
			Description: "Fix Service selector",
			Command:     "kubectl edit service database-hl",
			Notes:       "Change role=wrong-label to role=db",
		},
		{
			Description: "Verify endpoints",
			Command:     "kubectl get endpoints database-hl",
			Notes:       "Should now have IP addresses listed",
		},
	}
}
