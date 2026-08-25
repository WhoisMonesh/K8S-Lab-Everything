package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&ServiceWrongTargetPortLab{}) }

type ServiceWrongTargetPortLab struct{ BaseLab }

func (l *ServiceWrongTargetPortLab) ID() string          { return "service_wrong_targetport" }
func (l *ServiceWrongTargetPortLab) Title() string        { return "Service Points to Wrong TargetPort" }
func (l *ServiceWrongTargetPortLab) Category() Category   { return CategoryNetworking }
func (l *ServiceWrongTargetPortLab) Difficulty() Difficulty { return DifficultyEasy }
func (l *ServiceWrongTargetPortLab) EstimatedTime() int   { return 10 }
func (l *ServiceWrongTargetPortLab) Tags() []string {
	return []string{"service", "targetport", "endpoints", "networking"}
}
func (l *ServiceWrongTargetPortLab) Description() string {
	return `A Service named 'web-svc' is supposed to route traffic to pods running
nginx on port 80, but connections to the Service IP time out.

The Service has targetPort set to 8080, while the container actually
listens on port 80.

Your task: Fix the Service targetPort so traffic flows correctly.`
}
func (l *ServiceWrongTargetPortLab) Hints() []string {
	return []string{
		"Check kubectl get endpoints web-svc — is it empty or have addresses?",
		"Pod containers expose port 80 but Service targets 8080",
		"Fix: kubectl patch svc web-svc -p '{\"spec\":{\"ports\":[{\"targetPort\":80}]}}'",
	}
}

func (l *ServiceWrongTargetPortLab) Break(ctx context.Context, kp string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: web-pod
  namespace: default
  labels:
    app: web
spec:
  containers:
  - name: nginx
    image: nginx:1.27-alpine
    ports:
    - containerPort: 80
`
	svc := `apiVersion: v1
kind: Service
metadata:
  name: web-svc
  namespace: default
spec:
  selector:
    app: web
  ports:
  - port: 80
    targetPort: 8080
`
	kubectlApply(ctx, kp, pod)
	return kubectlApply(ctx, kp, svc)
}

func (l *ServiceWrongTargetPortLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(3 * time.Second)
	return nil
}

func (l *ServiceWrongTargetPortLab) Verify(ctx context.Context, kp string) error {
	tp, _ := kubectl(ctx, kp, "get", "svc", "web-svc", "-o",
		"jsonpath={.spec.ports[0].targetPort}")
	if tp == "8080" {
		return fmt.Errorf("targetPort is still 8080")
	}
	eps, _ := kubectl(ctx, kp, "get", "endpoints", "web-svc", "-o",
		"jsonpath={.subsets[0].addresses[0].ip}")
	if eps == "" {
		return fmt.Errorf("no endpoints — service still not matching pods")
	}
	return nil
}

func (l *ServiceWrongTargetPortLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check endpoints", Command: "kubectl get endpoints web-svc", Notes: "Empty or no addresses — mismatch"},
		{Description: "Check Service port", Command: "kubectl get svc web-svc -o yaml | grep targetPort", Notes: "Shows 8080 — wrong"},
		{Description: "Fix targetPort", Command: `kubectl patch svc web-svc -p '{"spec":{"ports":[{"targetPort":80}]}}'`, Notes: "Now matches containerPort 80"},
		{Description: "Verify", Command: "kubectl get endpoints web-svc && kubectl run curl --rm -it --image=curlimages/curl -- curl -s http://web-svc:80", Notes: "Endpoint has an address; curl returns nginx HTML"},
	}
}
