package labs

import (
	"context"
	"fmt"
	"strings"
)

func init() {
	Register(&CKADContainerImageBuildLab{})
}

type CKADContainerImageBuildLab struct {
	BaseLab
}

func (l *CKADContainerImageBuildLab) ID() string             { return "ckad_container_image_build" }
func (l *CKADContainerImageBuildLab) Title() string          { return "Build Container Image from Dockerfile" }
func (l *CKADContainerImageBuildLab) Category() Category     { return CategoryAppDesignBuild }
func (l *CKADContainerImageBuildLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKADContainerImageBuildLab) Cert() Cert             { return CertCKAD }
func (l *CKADContainerImageBuildLab) DomainWeight() int      { return 20 }
func (l *CKADContainerImageBuildLab) EstimatedTime() int     { return 20 }
func (l *CKADContainerImageBuildLab) Tags() []string {
	return []string{"dockerfile", "build", "image", "container"}
}

func (l *CKADContainerImageBuildLab) Description() string {
	return `A deployment references an image 'myapp:v1' that doesn't exist in the registry.
The pod is stuck in ImagePullBackOff.

Your task: Create a Dockerfile that builds a simple web server, build the image,
and update the deployment to use a valid image.`
}

func (l *CKADContainerImageBuildLab) Hints() []string {
	return []string{
		"Create a simple Dockerfile with a basic web server",
		"Build the image and load it into the cluster",
		"Update the deployment image reference",
	}
}

func (l *CKADContainerImageBuildLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKADContainerImageBuildLab) Break(ctx context.Context, kubeconfigPath string) error {
	deployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: myapp:v1
        ports:
        - containerPort: 8080`
	return kubectlApply(ctx, kubeconfigPath, deployment)
}

func (l *CKADContainerImageBuildLab) Verify(ctx context.Context, kubeconfigPath string) error {
	output, err := kubectl(ctx, kubeconfigPath, "get", "pods", "-l", "app=myapp",
		"-o", "jsonpath={.items[0].status.phase}")
	if err != nil {
		return fmt.Errorf("failed to get pod: %w", err)
	}
	if strings.TrimSpace(output) != "Running" {
		return fmt.Errorf("pod is not running (status: %s)", output)
	}
	return nil
}

func (l *CKADContainerImageBuildLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create Dockerfile", Command: "cat > Dockerfile <<EOF\nFROM nginx:alpine\nEXPOSE 8080\nEOF"},
		{Description: "Build image", Command: "docker build -t myapp:v1 ."},
		{Description: "Load image into cluster", Command: "kind load docker-image myapp:v1"},
		{Description: "Update deployment", Command: "kubectl set image deployment/myapp myapp=myapp:v1"},
	}
}
