package labs

import (
	"context"
)

func init() {
	Register(&CKSDockerfileSecurityLab{})
}

type CKSDockerfileSecurityLab struct {
	BaseLab
}

func (l *CKSDockerfileSecurityLab) ID() string             { return "cks_dockerfile_security" }
func (l *CKSDockerfileSecurityLab) Title() string          { return "Secure Dockerfile Best Practices" }
func (l *CKSDockerfileSecurityLab) Category() Category     { return CategorySupplyChain }
func (l *CKSDockerfileSecurityLab) Difficulty() Difficulty { return DifficultyMedium }
func (l *CKSDockerfileSecurityLab) EstimatedTime() int     { return 20 }
func (l *CKSDockerfileSecurityLab) Cert() Cert             { return CertCKS }
func (l *CKSDockerfileSecurityLab) DomainWeight() int      { return 20 }
func (l *CKSDockerfileSecurityLab) Tags() []string {
	return []string{"cks", "dockerfile", "security", "supply-chain"}
}

func (l *CKSDockerfileSecurityLab) Description() string {
	return `The Dockerfile for the application does not follow security best practices.
It runs as root, uses a full OS base image, and has unnecessary packages.

Your task: Create a secure Dockerfile that:
1. Uses a minimal base image (alpine or distroless)
2. Creates and uses a non-root user
3. Uses multi-stage build to reduce attack surface
4. Sets proper file permissions`
}

func (l *CKSDockerfileSecurityLab) Hints() []string {
	return []string{
		"Use FROM node:18-alpine as builder for build stage",
		"Create user with USER directive",
		"COPY --from=builder for final stage",
	}
}

func (l *CKSDockerfileSecurityLab) Prepare(ctx context.Context, kubeconfigPath string) error {
	return WaitForClusterReady(ctx, kubeconfigPath)
}

func (l *CKSDockerfileSecurityLab) Break(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSDockerfileSecurityLab) Verify(ctx context.Context, kubeconfigPath string) error {
	return nil
}

func (l *CKSDockerfileSecurityLab) SolutionSteps() []SolutionStep {
	return []SolutionStep{
		{Description: "Create secure Dockerfile", Command: `cat <<'EOF' > Dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .

FROM gcr.io/distroless/nodejs18-debian12
WORKDIR /app
COPY --from=builder /app .
USER nonroot:nonroot
EXPOSE 3000
CMD ["server.js"]
EOF`},
		{Description: "Build image", Command: "docker build -t secure-app:latest ."},
		{Description: "Verify user", Command: "docker run --rm secure-app:latest whoami"},
	}
}
