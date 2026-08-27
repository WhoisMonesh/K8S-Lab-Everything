package labs

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func init() {
	Register(&PodVolumeCSISecretLab{})
}

type PodVolumeCSISecretLab struct {
	BaseLab
}

func (l *PodVolumeCSISecretLab) ID() string {
	return "pod_volume_csi_secret"
}

func (l *PodVolumeCSISecretLab) Title() string {
	return "CSI Driver Secret Missing"
}

func (l *PodVolumeCSISecretLab) Category() Category {
	return CategoryStorage
}

func (l *PodVolumeCSISecretLab) Difficulty() Difficulty {
	return DifficultyHard
}

func (l *PodVolumeCSISecretLab) Description() string {
	return `A pod 'csi-app' is failing because it references a CSI secret that
doesn't exist. The CSI volume configuration requires credentials stored
in a Secret named 'csi-credentials' but the Secret is missing.

Your task: Create the missing Secret with the required CSI credentials.`
}

func (l *PodVolumeCSISecretLab) Hints() []string {
	return []string{
		"Check pod events for Secret not found errors",
		"The CSI volume driver requires a Secret",
		"Create the Secret with the correct CSI credentials",
	}
}

func (l *PodVolumeCSISecretLab) EstimatedTime() int {
	return 15
}

func (l *PodVolumeCSISecretLab) Tags() []string {
	return []string{"csi", "secret", "volume", "storage"}
}

func (l *PodVolumeCSISecretLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *PodVolumeCSISecretLab) Break(ctx context.Context, kubeconfigPath string) error {
	pod := `apiVersion: v1
kind: Pod
metadata:
  name: csi-app
  namespace: default
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ['sh', '-c', 'ls /data && sleep 3600']
    volumeMounts:
    - name: csi-vol
      mountPath: /data
  volumes:
  - name: csi-vol
    csi:
      driver: pd.csi.storage.gke.io
      volumeAttributes:
        diskName: test-disk
        diskType: pd-standard
      nodePublishSecretRef:
        name: csi-credentials
        namespace: default
`
	if err := kubectlApply(ctx, kubeconfigPath, pod); err != nil {
		return fmt.Errorf("creating pod: %w", err)
	}

	return nil
}

func (l *PodVolumeCSISecretLab) VerifyBroken(ctx context.Context, kubeconfigPath string) error {
	time.Sleep(10 * time.Second)
	return nil
}

func (l *PodVolumeCSISecretLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "secret", "csi-credentials",
		"-o", "jsonpath={.type}")
	if err != nil {
		return fmt.Errorf("failed to check secret: %w", err)
	}

	if strings.TrimSpace(output) != "Opaque" {
		return fmt.Errorf("Secret type is not correct")
	}

	output, err = kubectl(ctx, kubeconfigPath, "get", "pod", "csi-app",
		"-o", "jsonpath={.status.phase}")
	if err != nil {
		return fmt.Errorf("failed to check pod: %w", err)
	}

	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}

	return nil
}

func (l *PodVolumeCSISecretLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{
			Description: "Check pod events",
			Command:     "kubectl describe pod csi-app | grep -A 10 Events",
			Notes:       "Look for Secret 'csi-credentials' not found error",
		},
		{
			Description: "Create the missing Secret",
			Command:     "kubectl create secret generic csi-credentials --from-literal=username=admin --from-literal=password=secret123",
			Notes:       "Create Secret with required CSI credentials",
		},
		{
			Description: "Verify Secret exists",
			Command:     "kubectl get secret csi-credentials",
			Notes:       "Secret should now exist",
		},
		{
			Description: "Verify pod is running",
			Command:     "kubectl get pod csi-app",
			Notes:       "Pod should now be Running",
		},
	}
}
