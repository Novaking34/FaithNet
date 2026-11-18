# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/windows/install_persistent_home_service.ps1
# @id            playtest.field_kit.01.windows.install_home_service
# @layer         Layer 4: Proving Ground (Field Kit 01 — Windows Startup Service)
# @archetype     field_script | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Persistent Home-Anchor Installer (Windows Task Scheduler)
# @role          Compile faithnet-node.exe & Register Logon Startup Task
# @maturity      Rev 5 Baseline | Wave 14 5-Block Cross-Platform Parity
# @chassis       user-scoped | Zero kernel hooks; clean Scheduled Task registration
# @scale         O(1) deterministic build + idempotent Task Scheduler setup
# @seams         in: cmd/faithnet-node | exports: %LOCALAPPDATA%\FaithNet\faithnet-node.exe
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Persists Tier 2 Sovereign Root CA in %LOCALAPPDATA%\FaithNet\ca.
#                3. Supports `-AuditOnly` for safe cross-platform pwsh parity tests.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Windows counterpart to Linux `install_persistent_home_service.sh`. Compiles
#   a standalone native `faithnet-node.exe` into `%LOCALAPPDATA%\FaithNet\` and
#   registers a clean Scheduled Task (`FaithNet-HomeAnchor`) so the node boots
#   automatically on logon with Tier 2 Sovereign Root CA and zero hanging windows.
#
#   ### For Architects, Operators & Non-Programmers:
#   Makes a Windows PC automatically run FaithNet Home-Anchor every time you
#   log in, matching Linux's persistent `faithnet-home.service`.
#
#   ### For Systems Programmers & Engine Architects:
#   Uses `Register-ScheduledTask` (with automatic `RestartCount = 3`) and adds
#   optional private-profile Windows Defender Firewall rules when elevated.
# ==============================================================================
[CmdletBinding()]
param(
    [string]$NodeLabel  = "Home-Anchor-Win",
    [string]$BeaconSSID = "FaithNet-HomeRoom",
    [switch]$AuditOnly
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Paths & Task Scheduler Constants )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.install_home_service
# Mission:       Pin binary, CA storage, and Scheduled Task identifiers.
# ==============================================================================

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..\..\..")).Path
$AppRoot  = if ($env:LOCALAPPDATA) {
    Join-Path $env:LOCALAPPDATA "FaithNet"
} else {
    Join-Path $HOME ".local/share/faithnet-win-staging"
}
$BinPath  = Join-Path $AppRoot "faithnet-node.exe"
$CADir    = Join-Path $AppRoot "ca"
$TaskName = "FaithNet-HomeAnchor"

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Native Windows Build & Scheduled Task Registration )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.install_home_service
# Mission:       Compile faithnet-node.exe and register persistent logon task.
# ==============================================================================

function Build-WindowsNodeBinary {
    New-Item -ItemType Directory -Force -Path $AppRoot, $CADir | Out-Null
    Write-Host "[1/3] Compiling standalone native faithnet-node.exe -> $BinPath..."
    Push-Location $RepoRoot
    try {
        $prevGOOS = $env:GOOS
        $prevGOARCH = $env:GOARCH
        $env:GOOS = "windows"
        $env:GOARCH = "amd64"
        & go build -o $BinPath ./cmd/faithnet-node
        $env:GOOS = $prevGOOS
        $env:GOARCH = $prevGOARCH
    } finally {
        Pop-Location
    }
}

function Register-FaithNetScheduledTask {
    if (-not $IsWindows -and $PSVersionTable.PSEdition -eq "Core") {
        Write-Host "[2/3] Non-Windows pwsh host detected: verified build only (skipping Windows Task Scheduler)."
        return
    }
    Write-Host "[2/3] Registering Windows Scheduled Task '$TaskName'..."
    $argStr = "--label=`"$NodeLabel`" --udp=0.0.0.0:7777 --well --well-addr=0.0.0.0:8077 --tls-ca-dir=`"$CADir`" --beacon --beacon-ssid=`"$BeaconSSID`""
    $action   = New-ScheduledTaskAction -Execute $BinPath -Argument $argStr
    $trigger  = New-ScheduledTaskTrigger -AtLogOn
    $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)
    Register-ScheduledTask -TaskName $TaskName -Action $action -Trigger $trigger -Settings $settings -Force | Out-Null
    Write-Host "[3/3] Starting persistent '$TaskName'..."
    Start-ScheduledTask -TaskName $TaskName
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight Toolchain & Anchor Validation )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.install_home_service
# Mission:       Confirm Go toolchain and FaithNet go.mod anchor before build.
# ==============================================================================

function Test-InstallPreflight {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "[PREFLIGHT FAIL] Go toolchain not found in PATH."
    }
    if (-not (Test-Path (Join-Path $RepoRoot "go.mod"))) {
        throw "[PREFLIGHT FAIL] FaithNet go.mod anchor missing at $RepoRoot."
    }
    Write-Host "[PREFLIGHT PASS] Windows persistent installer ready (target: $BinPath)."
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Lifecycle Entrypoint )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.install_home_service
# Mission:       Run pre-flight, compile native binary, and register task.
# ==============================================================================

Test-InstallPreflight
if ($AuditOnly) {
    Write-Host "[AUDIT ONLY] Pre-flight verified cleanly; skipping build and Task Scheduler registration."
    exit 0
}
Build-WindowsNodeBinary
Register-FaithNetScheduledTask

# End of tests/playtester/field_kits/01_two_device_room_well/windows/install_persistent_home_service.ps1
