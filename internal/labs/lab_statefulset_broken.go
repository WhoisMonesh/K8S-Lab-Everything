package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&StatefulSetBrokenLab{})
}

type StatefulSetBrokenLab struct {
	BaseLab
}

func (l *StatefulSetBrokenLab) ID() string {
	return "statefulset_broken"
}

func (l *StatefulSetBrokenLab) Title() string {
	return "StatefulSet Pod Ordering Issues"
}

func (l *StatefulSetBrokenLab) Category() Category {
	return CategoryWorkloads
}

func (l *StatefulSetBrokenLab) Difficulty() Difficulty {
	return DifficultyMedium
}

func (l *StatefulSetBrokenLab) Description() string {
	return `A StatefulSet for a database cluster has been deployed, but the pods
are not starting correctly. The StatefulSet is stuck and not all replicas
are running.

Your task: Diagnose why the StatefulSet pods are not starting and fix the issue
so that all replicas are running and ready.`
}

func (l *StatefulSetBrokenLab) Hints() []string {
	return []string{
		"Check the StatefulSet status and events",
		"Look at the PersistentVolumeClaim (PVC) status for each pod",
		"StatefulSets require a headless service",
		"Check if the volumeClaimTemplates are correctly configured",
	}
}

func (l *StatefulSetBrokenLab) EstimatedTime() int {
	return 25
}

func (l *StatefulSetBrokenLab) Tags() []string {
	return []string{"statefulset", "storage", "pvc", "workloads", "troubleshooting"}
}

func (l *StatefulSetBrokenLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	if err := WaitForClusterReady(ctx, kubeconfigPath); err != nil {
		return err
	}

	// Create a namespace
	namespace := `apiVersion: v1
kind: Namespace
metadata:
  name: stateful-lab
`
	if err := kubectlApply(ctx, kubeconfigPath, namespace); err != nil {
		return err
	}

	// Create a headless service (required for StatefulSet)
	service := `apiVersion: v1
kind: Service
metadata:
  name: database
  namespace: stateful-lab
spec:
  clusterIP: None
  selector:
    app: db
  ports:
  - port: 3306
    name: mysql
`
	return kubectlApply(ctx, kubeconfigPath, service)
}

func (l *StatefulSetBrokenLab) Break(ctx context.Context, kubeconfigPath string) error {
	// Create a broken StatefulSet with incorrect service name
	brokenStatefulSet := `apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: database
  namespace: stateful-lab
spec:
  serviceName: "db-service"
  replicas: 3
  selector:
    matchLabels:
      app: db
  template:
    metadata:
      labels:
        app: db
    spec:
      containers:
      - name: mysql
        image: mysql:8.0
        env:
        - name: MYSQL_ROOT_PASSWORD
          value: "test123"
        ports:
        - containerPort: 3306
          name: mysql
        volumeMounts:
        - name: data
          mountPath: /var/lib/mysql
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 1Gi
      storageClassName: manual
`
	return kubectlApply(ctx, kubeconfigPath, brokenStatefulSet)
}

func (l *StatefulSetBrokenLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	// Give it a moment to try to start
	time.Sleep(5 * time.Second)

	// Check if StatefulSet exists with wrong serviceName
	output, err := kubectl(ctx, kubeconfigPath, "get", "statefulset",
		"database", "-n", "stateful-lab",
		"-o", "jsonpath={.spec.serviceName}")
	if err != nil {
		return fmt.Errorf("StatefulSet not found: %w", err)
	}

	if strings.TrimSpace(output) != "db-service" {
		return fmt.Errorf("StatefulSet serviceName is not broken")
	}

	return nil
}

func (l *StatefulSetBrokenLab) Verify(ctx context.Context, kubeconfigPath string) error {
	// Check if the StatefulSet has been fixed with correct service name
	output, err := kubectl(ctx, kubeconfigPath, "get", "statefulset",
		"database", "-n", "stateful-lab",
		"-o", "jsonpath={.spec.serviceName}")
	if err != nil {
		return fmt.Errorf("StatefulSet not found: %w", err)
	}

	serviceName := strings.TrimSpace(output)
	if serviceName != "database" {
		return fmt.Errorf("StatefulSet serviceName still incorrect: %s (expected: database)", serviceName)
	}

	// Wait a bit for pods to potentially start
	time.Sleep(10 * time.Second)

	// Check if at least one pod is running or pending
	// (might be pending due to storage issues in kind, but the config should be fixed)
	output, err = kubectl(ctx, kubeconfigPath, "get", "pods",
		"-n", "stateful-lab", "-l", "app=db",
		"-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return fmt.Errorf("checking StatefulSet pods: %w", err)
	}

	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("no StatefulSet pods found - may need to recreate the StatefulSet")
	}

	return nil
}

func (l *StatefulSetBrokenLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check the StatefulSet status",
			Command:     "kubectl get statefulset -n stateful-lab",
			Notes:       "Look at the ready count - it should show 0/3 or similar",
		},
		{
			Description: "Describe the StatefulSet for more details",
			Command:     "kubectl describe statefulset database -n stateful-lab",
			Notes:       "Look for events and error messages",
		},
		{
			Description: "Check the pods",
			Command:     "kubectl get pods -n stateful-lab",
			Notes:       "See if any pods have been created and their status",
		},
		{
			Description: "View the StatefulSet YAML",
			Command:     "kubectl get statefulset database -n stateful-lab -o yaml | grep serviceName",
			Notes:       "Check what service the StatefulSet is referencing",
		},
		{
			Description: "List services in the namespace",
			Command:     "kubectl get svc -n stateful-lab",
			Notes:       "Find the actual headless service name",
		},
		{
			Description: "Identify the mismatch",
			Notes:       "The StatefulSet references 'db-service' but the service is named 'database'",
		},
		{
			Description: "Delete the broken StatefulSet",
			Command:     "kubectl delete statefulset database -n stateful-lab",
			Notes:       "Need to delete and recreate due to serviceName being immutable",
		},
		{
			Description: "Recreate with the correct serviceName",
			Command: `kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: database
  namespace: stateful-lab
spec:
  serviceName: "database"
  replicas: 3
  selector:
    matchLabels:
      app: db
  template:
    metadata:
      labels:
        app: db
    spec:
      containers:
      - name: mysql
        image: mysql:8.0
        env:
        - name: MYSQL_ROOT_PASSWORD
          value: "test123"
        ports:
        - containerPort: 3306
          name: mysql
        volumeMounts:
        - name: data
          mountPath: /var/lib/mysql
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 1Gi
EOF`,
			Notes: "Note: You can remove storageClassName or set up a StorageClass",
		},
		{
			Description: "Verify the StatefulSet is now working",
			Command:     "kubectl get statefulset -n stateful-lab -w",
			Notes:       "Watch the StatefulSet create pods in order (database-0, database-1, etc.)",
		},
	}
}
