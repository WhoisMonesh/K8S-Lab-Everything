package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKSCICDSecurityLab{})
}

type CKSCICDSecurityLab struct {
	BaseLab
}

func (l *CKSCICDSecurityLab) ID() string             { return "cks_ci_cd_security" }
func (l *CKSCICDSecurityLab) Title() string          { return "Secure CI/CD Pipeline" }
func (l *CKSCICDSecurityLab) Category() Category     { return CategorySupplyChain }
func (l *CKSCICDSecurityLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSCICDSecurityLab) EstimatedTime() int     { return 30 }
func (l *CKSCICDSecurityLab) Cert() Cert             { return CertCKS }
func (l *CKSCICDSecurityLab) DomainWeight() int      { return 20 }
func (l *CKSCICDSecurityLab) Tags() []string {
	return []string{"cks", "ci-cd", "pipeline-security", "supply-chain"}
}

func (l *CKSCICDSecurityLab) Description() string {
	return `The CI/CD pipeline deploys images without scanning for vulnerabilities
or verifying image signatures. Images are pushed directly to production.

Your task: Create a Kubernetes CronJob that runs Trivy image scans
before deployments are allowed. The job should:
1. Scan images in the 'staging' namespace
2. Fail if critical vulnerabilities are found
3. Run on a daily schedule`
}

func (l *CKSCICDSecurityLab) Hints() []string {
	return []string{
		"Create a CronJob resource",
		"Use trivy image --severity CRITICAL --exit-code 1",
		"Set schedule with cron syntax",
	}
}

func (l *CKSCICDSecurityLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSCICDSecurityLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSCICDSecurityLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "cronjob", "-A", "-o", "name")
	if err != nil {
		return fmt.Errorf("failed to get cronjobs: %w", err)
	}
	if strings.Contains(output, "trivy-scan") {
		return nil
	}
	return fmt.Errorf("trivy scan CronJob not found")
}

func (l *CKSCICDSecurityLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create image scan CronJob", Command: `cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: CronJob
metadata:
  name: trivy-scan
  namespace: staging
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: trivy
            image: aquasec/trivy:latest
            args:
            - image
            - --severity
            - CRITICAL
            - --exit-code
            - "1"
            - nginx:latest
          restartPolicy: Never
EOF`},
		{Description: "Verify", Command: "kubectl get cronjob trivy-scan -n staging"},
	}
}
