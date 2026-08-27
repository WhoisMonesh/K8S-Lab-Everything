package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() { Register(&HostAliasWrongIPLab{}) }

type HostAliasWrongIPLab struct{ BaseLab }

func (l *HostAliasWrongIPLab) ID() string             { return "hostalias_wrong_ip" }
func (l *HostAliasWrongIPLab) Title() string          { return "Pod /etc/hosts Points to Wrong IP" }
func (l *HostAliasWrongIPLab) Category() Category     { return CategoryDNS }
func (l *HostAliasWrongIPLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *HostAliasWrongIPLab) EstimatedTime() int     { return 15 }
func (l *HostAliasWrongIPLab) Tags() []string {
	return []string{"dns", "hostalias", "hosts", "networking"}
}
func (l *HostAliasWrongIPLab) Description() string {
	return `A pod 'api-client' has a hostAliases entry pointing 'db.internal' to
192.168.1.999 (an invalid IP). The app crashes because it tries to
connect to this non-existent address.

Your task: Fix the hostAliases so 'db.internal' points to 10.96.0.1
(the cluster DNS service IP) and verify the pod can resolve it.`
}
func (l *HostAliasWrongIPLab) Hints() []string {
	return []string{
		"Check: kubectl get pod api-client -o yaml | grep -A5 hostAliases",
		"The IP 192.168.1.999 is invalid/unreachable",
		"Patch hostAliases to point db.internal → 10.96.0.1",
		"Pods are immutable for hostAliases — delete and recreate",
	}
}

func (l *HostAliasWrongIPLab) Break(ctx context.Context, kp string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: api-client
  namespace: default
spec:
  hostAliases:
  - ip: "192.168.1.999"
    hostnames:
    - "db.internal"
  containers:
  - name: app
    image: busybox:1.36
    command: ["sh","-c","getent hosts db.internal 2>&1; echo '---'; cat /etc/hosts | grep db; while true; do sleep 5; done"]
`
	return kubectlApply(ctx, kp, pod)
}

func (l *HostAliasWrongIPLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(3 * time.Second)
	return nil
}

func (l *HostAliasWrongIPLab) Verify(ctx context.Context, kp string) error {
	logs, _ := kubectl(ctx, kp, "logs", "api-client", "--tail=5")
	if strings.Contains(logs, "192.168.1.999") {
		return fmt.Errorf("hostAliases still points to 192.168.1.999")
	}
	if !strings.Contains(logs, "10.96.0.1") {
		return fmt.Errorf("db.internal not resolving to 10.96.0.1 — logs: %s", logs)
	}
	return nil
}

func (l *HostAliasWrongIPLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check current hosts", Command: "kubectl logs api-client --tail=5", Notes: "Shows 192.168.1.999 — wrong IP"},
		{Description: "Delete the pod (hostAliases is immutable)", Command: "kubectl delete pod api-client", Notes: "Must recreate with corrected spec"},
		{Description: "Recreate with correct IP", Command: `kubectl run api-client --image=busybox:1.36 --restart=Never --overrides='{"spec":{"hostAliases":[{"ip":"10.96.0.1","hostnames":["db.internal"]}]}}' -- sh -c "getent hosts db.internal; while true; do sleep 5; done"`, Notes: "db.internal now resolves to 10.96.0.1"},
		{Description: "Verify", Command: "kubectl logs api-client --tail=3", Notes: "Shows 10.96.0.1 db.internal"},
	}
}
