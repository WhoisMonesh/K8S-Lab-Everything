# K8S-Lab-Everything fully automated setup (Windows PowerShell 5.1+)
# Installs Go, Docker, kubectl, and a cluster provider (kind) automatically.
# Usage: powershell -ExecutionPolicy Bypass -File scripts\setup.ps1 [-ClusterProvider kind|k3d|minikube]

param(
    [ValidateSet("kind","k3d","minikube")]
    [string]$ClusterProvider = "kind"
)

$ErrorActionPreference = "Stop"

# ──────────────────────── Helpers ───────────────────────────────
function Test-Cmd {
    param([string]$Name)
    return [bool](Get-Command $Name -ErrorAction SilentlyContinue)
}

function Write-Info  { param([string]$Msg) Write-Host "[INFO]    $Msg" -ForegroundColor Cyan }
function Write-Ok    { param([string]$Msg) Write-Host "[OK]      $Msg" -ForegroundColor Green }
function Write-Warn  { param([string]$Msg) Write-Host "[WARN]    $Msg" -ForegroundColor Yellow }
function Write-Fail  { param([string]$Msg) Write-Host "[FAIL]    $Msg" -ForegroundColor Red }
function Write-Step  { param([string]$Msg) Write-Host "`n==> $Msg" -ForegroundColor Cyan }

# ──────────────────────── Ensure winget ─────────────────────────
function Ensure-Winget {
    if (Test-Cmd "winget") { return }
    Write-Info "Installing winget..."
    # winget comes with Windows 11 / modern Windows 10; try app installer
    try {
        Add-AppxPackage -RegisterByFamilyName -FamilyName Microsoft.DesktopAppInstaller_8wekyb3d8bbwe -ErrorAction SilentlyContinue
    } catch {
        Write-Warn "Could not auto-install winget. Please install it from the Microsoft Store."
    }
    if (-not (Test-Cmd "winget")) {
        throw "winget is required. Install from Microsoft Store and re-run."
    }
}

# ──────────────────────── Install Go ────────────────────────────
function Install-Go {
    if (Test-Cmd "go") {
        $v = (go version) -replace '.*go(\S+).*','$1'
        Write-Ok "Go $v already installed"
        return
    }

    Write-Step "Installing Go"
    Ensure-Winget

    try {
        winget install --id GoLang.Go --accept-source-agreements --accept-package-agreements -e
    } catch {
        Write-Warn "winget failed, downloading Go manually..."
        $goVer = "1.24.7"
        $url = "https://go.dev/dl/go${goVer}.windows-amd64.msi"
        $msiPath = "$env:TEMP\go-installer.msi"
        Invoke-WebRequest -Uri $url -OutFile $msiPath -UseBasicParsing
        Start-Process msiexec.exe -Wait -ArgumentList "/i `"$msiPath`" /quiet /norestart"
        Remove-Item $msiPath -Force -ErrorAction SilentlyContinue
    }

    # Refresh PATH
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

    if (Test-Cmd "go") {
        Write-Ok "Go installed"
    } else {
        Write-Warn "Go installed but not on PATH. Restart your terminal and re-run."
    }
}

# ──────────────────────── Install Docker ────────────────────────
function Install-Docker {
    if (Test-Cmd "docker") {
        Write-Ok "Docker already installed"
        return
    }

    Write-Step "Installing Docker Desktop"
    Ensure-Winget

    try {
        winget install --id Docker.DockerDesktop --accept-source-agreements --accept-package-agreements -e
    } catch {
        Write-Fail "Could not install Docker Desktop via winget."
        Write-Warn "Please install Docker Desktop manually from https://www.docker.com/products/docker-desktop"
        Write-Warn "After installation, open Docker Desktop and complete the setup wizard."
        Write-Host "Press Enter once Docker Desktop is running..." -ForegroundColor Yellow
        Read-Host
        return
    }

    Write-Warn "Docker Desktop installed. Please open Docker Desktop and complete the setup wizard."
    Write-Host "Press Enter once Docker Desktop is running..." -ForegroundColor Yellow
    Read-Host
    Write-Ok "Docker Desktop"
}

# ──────────────────────── Install kubectl ───────────────────────
function Install-Kubectl {
    if (Test-Cmd "kubectl") {
        Write-Ok "kubectl already installed"
        return
    }

    Write-Step "Installing kubectl"
    Ensure-Winget

    try {
        winget install --id Kubernetes.kubectl --accept-source-agreements --accept-package-agreements -e
    } catch {
        Write-Warn "winget failed, downloading kubectl manually..."
        $url = "https://dl.k8s.io/release/v1.30.0/bin/windows/amd64/kubectl.exe"
        $kubectlPath = "$env:LOCALAPPDATA\Microsoft\WinGet\Links\kubectl.exe"
        if (-not (Test-Path (Split-Path $kubectlPath))) {
            New-Item -ItemType Directory -Force -Path (Split-Path $kubectlPath) | Out-Null
        }
        Invoke-WebRequest -Uri $url -OutFile $kubectlPath -UseBasicParsing
    }

    # Refresh PATH
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

    if (Test-Cmd "kubectl") {
        Write-Ok "kubectl installed"
    } else {
        Write-Warn "kubectl installed but not on PATH. Restart your terminal."
    }
}

# ──────────────────────── Install cluster provider ──────────────
function Install-Kind {
    if (Test-Cmd "kind") {
        Write-Ok "kind already installed"
        return
    }

    Write-Step "Installing kind"
    Ensure-Winget

    try {
        winget install --id Kubernetes.kind --accept-source-agreements --accept-package-agreements -e
    } catch {
        Write-Warn "winget failed, downloading kind manually..."
        $url = "https://kind.sigs.k8s.io/dl/v0.25.0/kind-windows-amd64"
        $kindPath = "$env:LOCALAPPDATA\Microsoft\WinGet\Links\kind.exe"
        if (-not (Test-Path (Split-Path $kindPath))) {
            New-Item -ItemType Directory -Force -Path (Split-Path $kindPath) | Out-Null
        }
        Invoke-WebRequest -Uri $url -OutFile $kindPath -UseBasicParsing
    }

    # Refresh PATH
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

    if (Test-Cmd "kind") {
        Write-Ok "kind installed"
    } else {
        Write-Warn "kind installed but not on PATH. Restart your terminal."
    }
}

function Install-K3d {
    if (Test-Cmd "k3d") {
        Write-Ok "k3d already installed"
        return
    }

    Write-Step "Installing k3d"
    Ensure-Winget

    try {
        winget install --id K3D.k3d --accept-source-agreements --accept-package-agreements -e
    } catch {
        Write-Warn "winget failed, downloading k3d manually..."
        $url = "https://github.com/k3d-io/k3d/releases/latest/download/k3d-windows-amd64.exe"
        $k3dPath = "$env:LOCALAPPDATA\Microsoft\WinGet\Links\k3d.exe"
        if (-not (Test-Path (Split-Path $k3dPath))) {
            New-Item -ItemType Directory -Force -Path (Split-Path $k3dPath) | Out-Null
        }
        Invoke-WebRequest -Uri $url -OutFile $k3dPath -UseBasicParsing
    }

    # Refresh PATH
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

    if (Test-Cmd "k3d") {
        Write-Ok "k3d installed"
    } else {
        Write-Warn "k3d installed but not on PATH. Restart your terminal."
    }
}

function Install-Minikube {
    if (Test-Cmd "minikube") {
        Write-Ok "minikube already installed"
        return
    }

    Write-Step "Installing minikube"
    Ensure-Winget

    try {
        winget install --id Kubernetes.minikube --accept-source-agreements --accept-package-agreements -e
    } catch {
        Write-Warn "winget failed, downloading minikube manually..."
        $url = "https://storage.googleapis.com/minikube/releases/latest/minikube-windows-amd64.exe"
        $miniPath = "$env:LOCALAPPDATA\Microsoft\WinGet\Links\minikube.exe"
        if (-not (Test-Path (Split-Path $miniPath))) {
            New-Item -ItemType Directory -Force -Path (Split-Path $miniPath) | Out-Null
        }
        Invoke-WebRequest -Uri $url -OutFile $miniPath -UseBasicParsing
    }

    # Refresh PATH
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

    if (Test-Cmd "minikube") {
        Write-Ok "minikube installed"
    } else {
        Write-Warn "minikube installed but not on PATH. Restart your terminal."
    }
}

function Install-ClusterProvider {
    switch ($ClusterProvider) {
        "kind"     { Install-Kind }
        "k3d"      { Install-K3d }
        "minikube" { Install-Minikube }
    }
}

# ──────────────────────── Build binary ──────────────────────────
function Build-Binary {
    if (-not (Test-Path ".\go.mod")) {
        Write-Warn "go.mod not found — skipping binary build"
        return
    }

    if (-not (Test-Cmd "go")) {
        Write-Warn "Go not available — cannot build binary"
        return
    }

    Write-Step "Building cka-lab-runner"
    if (-not (Test-Path ".\bin")) { New-Item -ItemType Directory -Path ".\bin" | Out-Null }
    go build -o .\bin\cka-lab-runner.exe -v .\cmd\cka-lab-runner
    Write-Ok "Binary built: bin\cka-lab-runner.exe"
}

# ──────────────────────── Main ──────────────────────────────────
Write-Host ""
Write-Host "══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  K8S-Lab-Everything — Automated Setup (Windows)" -ForegroundColor Cyan
Write-Host "══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Info "Cluster provider: $ClusterProvider"
Write-Host ""

Install-Go
Install-Docker
Install-Kubectl
Install-ClusterProvider
Build-Binary

Write-Host ""
Write-Host "══════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host "  All prerequisites installed!" -ForegroundColor Green
Write-Host "══════════════════════════════════════════════════════════" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:"
Write-Host "  .\bin\cka-lab-runner.exe init"
Write-Host "  .\bin\cka-lab-runner.exe up"
Write-Host "  .\bin\cka-lab-runner.exe lab list"
Write-Host "  .\bin\cka-lab-runner.exe lab run pod_crashloop"
Write-Host ""
