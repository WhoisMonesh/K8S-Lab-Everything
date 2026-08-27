#Requires -Version 5.1
<#
.SYNOPSIS
    Uninstalls cka-lab-runner from Windows.

.DESCRIPTION
    Removes the cka-lab-runner binary, config files, and progress data.

.PARAMETER DryRun
    Show what would be removed without deleting anything.

.PARAMETER Force
    Skip confirmation prompt.
#>
param(
    [switch]$DryRun,
    [switch]$Force
)

$ErrorActionPreference = "SilentlyContinue"

# Helper functions
function Write-Status {
    param([string]$Message, [string]$Color = "White")
    Write-Host "  " -NoNewline
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success {
    param([string]$Message)
    Write-Host "  " -NoNewline
    Write-Host "✔ " -ForegroundColor Green -NoNewline
    Write-Host $Message
}

function Write-Fail {
    param([string]$Message)
    Write-Host "  " -NoNewline
    Write-Host "✖ " -ForegroundColor Red -NoNewline
    Write-Host $Message
}

# Banner
Write-Host ""
Write-Host "  ┌─────────────────────────────────────────────────────────────┐" -ForegroundColor Cyan
Write-Host "  │" -ForegroundColor Cyan -NoNewline
Write-Host "  K8S-Lab-Everything — Uninstaller" -ForegroundColor White -NoNewline
Write-Host "                          │" -ForegroundColor Cyan
Write-Host "  └─────────────────────────────────────────────────────────────┘" -ForegroundColor Cyan
Write-Host ""

# Find items to remove
$Items = @()

# Binary paths
$BinaryPaths = @(
    "$env:USERPROFILE\go\bin\cka-lab-runner.exe",
    "$env:LOCALAPPDATA\Microsoft\WinGet\Links\cka-lab-runner.exe",
    "$env:ProgramFiles\cka-lab-runner\cka-lab-runner.exe"
)

foreach ($path in $BinaryPaths) {
    if (Test-Path $path) {
        $Items += $path
    }
}

# Also check PATH for cka-lab-runner
$cmd = Get-Command cka-lab-runner -ErrorAction SilentlyContinue
if ($cmd -and $cmd.Source -and (Test-Path $cmd.Source)) {
    $source = $cmd.Source
    if ($Items -notcontains $source) {
        $Items += $source
    }
}

# Config directory
$ConfigDir = "$env:USERPROFILE\.cka-lab-runner"
if (Test-Path $ConfigDir) {
    $Items += $ConfigDir
}

# AppData
$AppDataDir = "$env:APPDATA\cka-lab-runner"
if (Test-Path $AppDataDir) {
    $Items += $AppDataDir
}

# Local progress/config files in current directory
$LocalFiles = @(
    ".lab-progress.json",
    "cka-lab-runner.yaml"
)
foreach ($file in $LocalFiles) {
    if (Test-Path $file) {
        $Items += (Resolve-Path $file).Path
    }
}

# Nothing found
if ($Items.Count -eq 0) {
    Write-Host ""
    Write-Success "Nothing to remove — cka-lab-runner is not installed"
    Write-Host ""
    exit 0
}

# Show what will be removed
Write-Host ""
Write-Status "Items to remove:" "White"
Write-Host ""

foreach ($item in $Items) {
    if (Test-Path $item -PathType Leaf) {
        Write-Host "    " -NoNewline
        Write-Host "file   " -ForegroundColor DarkGray -NoNewline
        Write-Host $item
    }
    elseif (Test-Path $item -PathType Container) {
        Write-Host "    " -NoNewline
        Write-Host "dir    " -ForegroundColor DarkGray -NoNewline
        Write-Host $item
    }
}
Write-Host ""

# Dry run
if ($DryRun) {
    Write-Status "Dry run — nothing was deleted" "Yellow"
    Write-Host ""
    exit 0
}

# Confirm
if (-not $Force) {
    Write-Host "  " -NoNewline
    Write-Host "⚠  Remove all listed items? [y/N] " -ForegroundColor Yellow -NoNewline
    $answer = Read-Host
    if ($answer -notmatch "^[Yy]$") {
        Write-Host ""
        Write-Status "Cancelled." "DarkGray"
        Write-Host ""
        exit 0
    }
}

Write-Host ""

# Remove items
foreach ($item in $Items) {
    try {
        if (Test-Path $item -PathType Leaf) {
            Remove-Item -Path $item -Force -ErrorAction Stop
            Write-Success "Removed $item"
        }
        elseif (Test-Path $item -PathType Container) {
            Remove-Item -Path $item -Recurse -Force -ErrorAction Stop
            Write-Success "Removed $item"
        }
    }
    catch {
        Write-Fail "Failed to remove $item (run as Administrator?)"
    }
}

Write-Host ""
Write-Host "  " -NoNewline
Write-Host "✔ " -ForegroundColor Green -NoNewline
Write-Host "cka-lab-runner has been uninstalled" -ForegroundColor White
Write-Host ""
