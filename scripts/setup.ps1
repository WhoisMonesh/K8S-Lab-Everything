# K8S-Lab-Everything prerequisite checker (Windows PowerShell 5.1+)
# Usage: powershell -ExecutionPolicy Bypass -File scripts\setup.ps1

$ErrorActionPreference = "Continue"

function Test-Cmd {
    param([string]$Name)
    return [bool](Get-Command $Name -ErrorAction SilentlyContinue)
}

$missing = 0
Write-Host ""
Write-Host "== K8S-Lab-Everything setup check ==" -ForegroundColor Cyan
Write-Host ""

if (Test-Cmd "go") {
    Write-Host "[OK]       go" -ForegroundColor Green
} else {
    Write-Host "[MISSING]  go" -ForegroundColor Red
    Write-Host "           Install: winget install GoLang.Go" -ForegroundColor Yellow
    $script:missing++
}

if (Test-Cmd "docker") {
    Write-Host "[OK]       docker" -ForegroundColor Green
} else {
    Write-Host "[MISSING]  docker" -ForegroundColor Red
    Write-Host "           Install: winget install Docker.DockerDesktop (then start Docker Desktop)" -ForegroundColor Yellow
    $script:missing++
}

if (Test-Cmd "kubectl") {
    Write-Host "[OK]       kubectl" -ForegroundColor Green
} else {
    Write-Host "[MISSING]  kubectl" -ForegroundColor Red
    Write-Host "           Install: winget install Kubernetes.kubectl" -ForegroundColor Yellow
    $script:missing++
}

if ((Test-Cmd "kind") -or (Test-Cmd "k3d") -or (Test-Cmd "minikube")) {
    Write-Host "[OK]       cluster provider (kind / k3d / minikube)" -ForegroundColor Green
} else {
    Write-Host "[MISSING]  cluster provider - install at least one:" -ForegroundColor Red
    Write-Host "             winget install Kubernetes.kind" -ForegroundColor Yellow
    Write-Host "             winget install K3D.k3d" -ForegroundColor Yellow
    Write-Host "             winget install Kubernetes.minikube" -ForegroundColor Yellow
    $script:missing++
}

Write-Host ""

if ($missing -eq 0) {
    Write-Host "All prerequisites installed! Next steps:" -ForegroundColor Green
    Write-Host "  go install ./cmd/cka-lab-runner"
    Write-Host "  cka-lab-runner init"
    Write-Host "  cka-lab-runner up"
    Write-Host "  cka-lab-runner lab list"
} else {
    Write-Host "$missing prerequisite(s) missing - install them, then re-run this script." -ForegroundColor Yellow
}
Write-Host ""
