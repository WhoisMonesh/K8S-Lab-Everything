package labs

import (
	"context"
)

func init() {
	Register(&CKSPrivilegeEscalationDetectLab{})
}

type CKSPrivilegeEscalationDetectLab struct {
	BaseLab
}

func (l *CKSPrivilegeEscalationDetectLab) ID() string { return "cks_privilege_escalation_detect" }
func (l *CKSPrivilegeEscalationDetectLab) Title() string {
	return "Detect Privilege Escalation Attempts"
}
func (l *CKSPrivilegeEscalationDetectLab) Category() Category     { return CategoryMonitoringLogging }
func (l *CKSPrivilegeEscalationDetectLab) Difficulty() Difficulty { return DifficultyHard }
func (l *CKSPrivilegeEscalationDetectLab) EstimatedTime() int     { return 30 }
func (l *CKSPrivilegeEscalationDetectLab) Cert() Cert             { return CertCKS }
func (l *CKSPrivilegeEscalationDetectLab) DomainWeight() int      { return 20 }
func (l *CKSPrivilegeEscalationDetectLab) Tags() []string {
	return []string{"cks", "privilege-escalation", "detection", "monitoring"}
}

func (l *CKSPrivilegeEscalationDetectLab) Description() string {
	return `Privilege escalation attempts such as binding to cluster-admin, creating
ClusterRoleBindings, or running privileged containers are not being monitored.

Your task: Create a Falco rule that detects:
1. Creating ClusterRoleBindings
2. Running privileged containers
3. Exec into system pods`
}

func (l *CKSPrivilegeEscalationDetectLab) Hints() []string {
	return []string{
		"Create a Falco custom rule",
		"Use Kubernetes audit events as data source",
		"Alert on cluster-admin binding creation",
	}
}

func (l *CKSPrivilegeEscalationDetectLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSPrivilegeEscalationDetectLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSPrivilegeEscalationDetectLab) Verify(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSPrivilegeEscalationDetectLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create Falco custom rules", Command: `cat <<EOF > /etc/falco/rules.d/privilege-escalation.yaml
- rule: Detect ClusterRoleBinding Creation
  desc: Detect creation of ClusterRoleBindings which may indicate privilege escalation
  condition: >
    ke.become.clusterrolebinding and not ka.target.name in (allowed_clusterrolebindings)
  output: >
    ClusterRoleBinding created (user=%ka.user.name binding=%ka.target.name)
  priority: WARNING
  tags: [k8s, privilege-escalation]

- rule: Detect Privileged Container
  desc: Detect running privileged containers
  condition: >
    container and proc.name in (privileged_procs) and not container.image.repository in (trusted_repos)
  output: >
    Privileged container run (user=%user.name container=%container.name image=%container.image.repository)
  priority: CRITICAL
  tags: [k8s, privilege-escalation, container]
EOF`},
		{Description: "Reload Falco rules", Command: "kubectl exec -n falco falco-xxxxx -- falcosidekick --reload"},
	}
}
