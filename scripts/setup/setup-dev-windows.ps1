# CLASSIFICATION: UNCLASSIFIED
# RTSA Windows Bootstrap Script
# Sets up WSL2 + Ubuntu and launches the Linux dev setup script.
# Run as Administrator in PowerShell: .\scripts\setup\setup-dev-windows.ps1

#Requires -RunAsAdministrator

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$UBUNTU_DISTRO = 'Ubuntu-24.04'
$REPO_ROOT = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

function Write-Step { Write-Host "`n== $args ==" -ForegroundColor Cyan }
function Write-Ok { Write-Host "[OK] $args"   -ForegroundColor Green }
function Write-Warn { Write-Host "[!]  $args"   -ForegroundColor Yellow }
function Write-Err { Write-Host "[X]  $args"   -ForegroundColor Red }

# ─────────────────────────────────────────────────────────────
# Step 1: Enable WSL2
# ─────────────────────────────────────────────────────────────
function Enable-WSL {
    Write-Step 'Enabling WSL2'

    $wsl = Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Windows-Subsystem-Linux
    if ($wsl.State -ne 'Enabled') {
        Write-Host 'Enabling WSL feature...'
        Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Windows-Subsystem-Linux -NoRestart
    }
    else {
        Write-Ok 'WSL feature already enabled'
    }

    $vmp = Get-WindowsOptionalFeature -Online -FeatureName VirtualMachinePlatform
    if ($vmp.State -ne 'Enabled') {
        Write-Host 'Enabling Virtual Machine Platform...'
        Enable-WindowsOptionalFeature -Online -FeatureName VirtualMachinePlatform -NoRestart
    }
    else {
        Write-Ok 'Virtual Machine Platform already enabled'
    }

    # Set WSL2 as default
    wsl --set-default-version 2 2>$null
    Write-Ok 'WSL2 set as default version'
}

# ─────────────────────────────────────────────────────────────
# Step 2: Install Ubuntu distribution
# ─────────────────────────────────────────────────────────────
function Install-Ubuntu {
    Write-Step "Installing $UBUNTU_DISTRO"

    $installed = wsl --list --quiet 2>$null | Where-Object { $_ -match [regex]::Escape($UBUNTU_DISTRO) }
    if ($installed) {
        Write-Ok "$UBUNTU_DISTRO already installed"
        return
    }

    Write-Host "Installing ${UBUNTU_DISTRO} via wsl --install ..."
    wsl --install -d $UBUNTU_DISTRO

    Write-Warn "A restart may be required. If Ubuntu setup window did not appear, restart and re-run this script."
}

# ─────────────────────────────────────────────────────────────
# Step 3: Verify WSL Ubuntu is running
# ─────────────────────────────────────────────────────────────
function Test-UbuntuRunning {
    Write-Step "Verifying $UBUNTU_DISTRO is running"

    $state = wsl -d $UBUNTU_DISTRO --exec echo 'ready' 2>$null
    if ($state -eq 'ready') {
        Write-Ok "$UBUNTU_DISTRO is running"
        return $true
    }
    else {
        Write-Warn "$UBUNTU_DISTRO is not yet initialized. Open the Ubuntu app to complete setup, then re-run this script."
        return $false
    }
}

# ─────────────────────────────────────────────────────────────
# Step 4: Install Docker Desktop
# ─────────────────────────────────────────────────────────────
function Install-DockerDesktop {
    Write-Step 'Checking Docker Desktop'

    if (Get-Command docker -ErrorAction SilentlyContinue) {
        Write-Ok "Docker already installed: $(docker --version)"
        return
    }

    Write-Warn 'Docker Desktop not found.'
    Write-Host 'Please download and install Docker Desktop from:'
    Write-Host '  https://www.docker.com/products/docker-desktop/'
    Write-Host ''
    Write-Host 'After installation:'
    Write-Host '  1. Open Docker Desktop'
    Write-Host '  2. Go to Settings → Resources → WSL Integration'
    Write-Host "  3. Enable integration for $UBUNTU_DISTRO"
    Write-Host '  4. Re-run this script'

    $open = Read-Host 'Open Docker Desktop download page now? (y/N)'
    if ($open -eq 'y' -or $open -eq 'Y') {
        Start-Process 'https://www.docker.com/products/docker-desktop/'
    }
}

# ─────────────────────────────────────────────────────────────
# Step 5: Configure VS Code WSL extension
# ─────────────────────────────────────────────────────────────
function Configure-VSCode {
    Write-Step 'Checking VS Code WSL extension'

    if (Get-Command code -ErrorAction SilentlyContinue) {
        $extensions = code --list-extensions 2>$null
        if ($extensions -match 'ms-vscode-remote.remote-wsl') {
            Write-Ok 'VS Code WSL extension already installed'
        }
        else {
            Write-Host 'Installing VS Code WSL extension...'
            code --install-extension ms-vscode-remote.remote-wsl
            Write-Ok 'VS Code WSL extension installed'
        }
    }
    else {
        Write-Warn 'VS Code not found. Install from: https://code.visualstudio.com/'
    }
}

# ─────────────────────────────────────────────────────────────
# Step 6: Run Linux setup script inside WSL2
# ─────────────────────────────────────────────────────────────
function Invoke-LinuxSetup {
    Write-Step 'Running Linux developer setup inside WSL2'

    # Convert Windows path to WSL path
    $wsl_repo = $REPO_ROOT -replace '\\', '/' -replace '^([A-Z]):', { "/mnt/$($_.Value[0].ToString().ToLower())" }

    Write-Host "Repository path in WSL: $wsl_repo"
    Write-Host ''

    # Run the setup script inside the Ubuntu WSL distribution
    wsl -d $UBUNTU_DISTRO -- bash -c "
        set -e
        cd '$wsl_repo'
        chmod +x scripts/setup/setup-dev.sh scripts/setup/gen-dev-certs.sh scripts/dev/*.sh 2>/dev/null || true
        ./scripts/setup/setup-dev.sh
    "
}

# ─────────────────────────────────────────────────────────────
# Step 7: Print next steps
# ─────────────────────────────────────────────────────────────
function Write-NextSteps {
    Write-Host ''
    Write-Host '════════════════════════════════════════════════════' -ForegroundColor Cyan
    Write-Host '  Windows Setup Complete' -ForegroundColor Green
    Write-Host ''
    Write-Host '  Next steps (run inside WSL2 Ubuntu terminal):' -ForegroundColor Cyan
    Write-Host ''
    Write-Host "  1. Open WSL: wsl -d $UBUNTU_DISTRO"
    Write-Host "  2. Navigate to repo:"
    $wsl_repo = $REPO_ROOT -replace '\\', '/' -replace '^([A-Z]):', { "/mnt/$($_.Value[0].ToString().ToLower())" }
    Write-Host "     cd $wsl_repo"
    Write-Host '  3. Start dev stack:'
    Write-Host '     docker compose -f deploy/docker-compose.yml up -d'
    Write-Host '  4. Initialize services:'
    Write-Host '     ./scripts/dev/init-topics.sh'
    Write-Host '     ./scripts/dev/init-clickhouse.sh'
    Write-Host '  5. Verify setup:'
    Write-Host '     ./scripts/dev/health-check.sh'
    Write-Host ''
    Write-Host '  Or open VS Code in WSL:'
    Write-Host "  wsl -d $UBUNTU_DISTRO -- code $wsl_repo"
    Write-Host '════════════════════════════════════════════════════' -ForegroundColor Cyan
}

# ─────────────────────────────────────────────────────────────
# Main
# ─────────────────────────────────────────────────────────────
Write-Host ''
Write-Host '╔════════════════════════════════════════════════════╗' -ForegroundColor Cyan
Write-Host '║  RTSA Windows Developer Bootstrap                  ║' -ForegroundColor Cyan
Write-Host '║  CLASSIFICATION: UNCLASSIFIED                      ║' -ForegroundColor Cyan
Write-Host '╚════════════════════════════════════════════════════╝' -ForegroundColor Cyan
Write-Host ''

Enable-WSL
Install-Ubuntu
Install-DockerDesktop
Configure-VSCode

if (Test-UbuntuRunning) {
    Invoke-LinuxSetup
}
else {
    Write-Warn 'Skipping Linux setup — Ubuntu not yet initialized.'
    Write-Host 'After Ubuntu setup is complete, re-run this script.'
}

Write-NextSteps
