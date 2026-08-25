package labs

import (
	"context"
	"fmt"
	"time"
)

func init() { Register(&VolumeSubpathMissingLab{}) }

type VolumeSubpathMissingLab struct{ BaseLab }

func (l *VolumeSubpathMissingLab) ID() string          { return "volume_subpath_missing" }
func (l *VolumeSubpathMissingLab) Title() string        { return "Pod CrashLoop — Wrong Volume subPath" }
func (l *VolumeSubpathMissingLab) Category() Category   { return CategoryStorage }
func (l *VolumeSubpathMissingLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *VolumeSubpathMissingLab) EstimatedTime() int   { return 15 }
func (l *VolumeSubpathMissingLab) Tags() []string {
	return []string{"volume", "subpath", "configmap", "storage"}
}
func (l *VolumeSubpathMissingLab) Description() string {
	return `A pod 'config-app' is in CrashLoopBackOff. It mounts a ConfigMap
'nginx-conf' into /etc/nginx/nginx.conf, but the ConfigMap's data key
is 'default.conf' not 'nginx.conf'.

The volume subPath is set to 'nginx.conf' which doesn't exist as a
key in the ConfigMap.

Your task: Fix the subPath to match the actual ConfigMap data key so
the file mounts correctly.`
}
func (l *VolumeSubpathMissingLab) Hints() []string {
	return []string{
		"Check: kubectl get cm nginx-conf -o yaml — see the data keys",
		"The key is 'default.conf' but subPath is set to 'nginx.conf'",
		"Fix: kubectl patch deploy config-app -p '{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"nginx\",\"volumeMounts\":[{\"subPath\":\"default.conf\"}]}]}}}}'",
	}
}

func (l *VolumeSubpathMissingLab) Break(ctx context.Context, kp string) error {
	cm := `apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-conf
  namespace: default
data:
  default.conf: |
    server {
      listen 80;
      server_name localhost;
      location / { return 200 "ok"; }
    }
`
	deploy := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: config-app
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: config-app
  template:
    metadata:
      labels:
        app: config-app
    spec:
      containers:
      - name: nginx
        image: nginx:1.27-alpine
        volumeMounts:
        - name: config
          mountPath: /etc/nginx/nginx.conf
          subPath: nginx.conf
      volumes:
      - name: config
        configMap:
          name: nginx-conf
`
	kubectlApply(ctx, kp, cm)
	return kubectlApply(ctx, kp, deploy)
}

func (l *VolumeSubpathMissingLab) VerifyBroken(_ context.Context, _ string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *VolumeSubpathMissingLab) Verify(ctx context.Context, kp string) error {
	subPath, _ := kubectl(ctx, kp, "get", "deploy", "config-app", "-o",
		"jsonpath={.spec.template.spec.containers[0].volumeMounts[0].subPath}")
	if subPath == "nginx.conf" {
		return fmt.Errorf("subPath is still nginx.conf — doesn't match ConfigMap key")
	}
	ready, _ := kubectl(ctx, kp, "get", "deploy", "config-app", "-o",
		"jsonpath={.status.readyReplicas}")
	if ready != "1" {
		return fmt.Errorf("deployment not ready (ready: %s)", ready)
	}
	return nil
}

func (l *VolumeSubpathMissingLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Check ConfigMap keys", Command: "kubectl get cm nginx-conf -o jsonpath='{.data}'", Notes: "Key is 'default.conf', not 'nginx.conf'"},
		{Description: "Fix subPath", Command: `kubectl patch deploy config-app --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/volumeMounts/0/subPath","value":"default.conf"}]'`, Notes: "SubPath must match the ConfigMap data key"},
		{Description: "Verify", Command: "kubectl rollout status deploy/config-app && kubectl exec deploy/config-app -- cat /etc/nginx/nginx.conf", Notes: "Nginx config is correctly mounted"},
	}
}
