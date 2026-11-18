# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/windows/run_room_well_host.ps1
# @id            playtest.field_kit.01.windows.room_well_host
# @layer         Layer 4: Proving Ground (Field Kit 01 — Windows Room & Well)
# @archetype     field_script | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Field Kit 01 — Two-Device Room & Well Launcher (Windows)
# @role          Foreground Room & Well Node Launcher (PowerShell 5.1 / 7+)
# @maturity      Rev 5 Baseline | Wave 14 5-Block Cross-Platform Parity
# @chassis       stateless-wrapper | Zero permanent registry/firewall changes
# @scale         O(1) host LAN IP discovery + foreground faithnet-node launch
# @seams         in: cmd/faithnet-node | exports: live UDP :7777 & Well :8077
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Parseable & testable via PowerShell 7 (`pwsh`) on Linux too.
#                3. Supports `-AuditOnly` for safe non-mutating parity checks.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Windows PowerShell counterpart to Linux `run_room_well_host.sh`. Resolves
#   the host's active LAN IPv4 address, prints visitor Well/CA/pairing URLs,
#   and runs `faithnet-node` with Tier 2 Persistent Sovereign Local Root CA.
#
#   ### For Architects, Operators & Non-Programmers:
#   Turns any Windows 10/11 PC into a live FaithNet local room anchor in one
#   PowerShell command so phones and nearby laptops can connect over Wi-Fi.
#
#   ### For Systems Programmers & Engine Architects:
#   Uses cross-platform .NET `[System.Net.NetworkInformation.NetworkInterface]`
#   so both Windows PowerShell 5.1 and PowerShell 7 (`pwsh` on Linux) execute
#   identical pre-flight and IP discovery logic cleanly.
# ==============================================================================
[CmdletBinding()]
param(
    [string]$NodeLabel  = "Home-Anchor-Win",
    [string]$UDPAddr    = "0.0.0.0:7777",
    [string]$WellAddr   = "0.0.0.0:8077",
    [string]$BeaconSSID = "FaithNet-HomeRoom",
    [string]$TLSCADir   = "",
    [switch]$Interactive,
    [switch]$AuditOnly
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Paths, Defaults & Cross-Platform LAN IPv4 Resolver )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.room_well_host
# Mission:       Resolve FaithNet repository root, CA directory, and LAN IP.
# ==============================================================================

$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..\..\..")).Path
if ([string]::IsNullOrWhiteSpace($TLSCADir)) {
    $BaseTemp = if ($env:TEMP) { $env:TEMP } else { [System.IO.Path]::GetTempPath() }
    $TLSCADir = Join-Path $BaseTemp "faithnet-sovereign-ca"
}

function Get-HostLANIPv4 {
    try {
        $ifaces = [System.Net.NetworkInformation.NetworkInterface]::GetAllNetworkInterfaces() |
            Where-Object {
                $_.OperationalStatus -eq [System.Net.NetworkInformation.OperationalStatus]::Up -and
                $_.NetworkInterfaceType -ne [System.Net.NetworkInformation.NetworkInterfaceType]::Loopback
            }
        foreach ($nic in $ifaces) {
            foreach ($u in $nic.GetIPProperties().UnicastAddresses) {
                if ($u.Address.AddressFamily -eq [System.Net.Sockets.AddressFamily]::InterNetwork) {
                    $ipStr = $u.Address.IPAddressToString
                    if (-not $ipStr.StartsWith("169.254.")) {
                        return $ipStr
                    }
                }
            }
        }
    } catch { }
    return "127.0.0.1"
}

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Visitor Onboarding Banner & Sovereign Node Execution )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.room_well_host
# Mission:       Display visitor endpoints and start faithnet-node cleanly.
# ==============================================================================

function Show-VisitorBanner {
    param([string]$HostIP)
    Write-Host "=== FAITHNET FIELD KIT 01: TWO-DEVICE ROOM & THE WELL (WINDOWS) ==="
    Write-Host "Host LAN IP : $HostIP"
    Write-Host "The Well URL: http://${HostIP}:8077/  (or https://${HostIP}:8077/)"
    Write-Host "LAN UDP Port: ${HostIP}:7777"
    Write-Host "Root CA Cert: http://${HostIP}:8077/faithnet-root-ca.crt"
    Write-Host "Win AutoPair: irm http://${HostIP}:8077/pair.ps1 | iex"
    Write-Host "Nix AutoPair: curl -fsSL http://${HostIP}:8077/pair.sh | sh"
    Write-Host ""
    Write-Host "NOTE (Windows Defender Firewall): If prompted, allow Private network access on TCP 8077 & UDP 7777."
    Write-Host ""
}

function Invoke-RoomWellNode {
    $argsList = @(
        "run", "./cmd/faithnet-node",
        "--label=$NodeLabel",
        "--udp=$UDPAddr",
        "--well",
        "--well-addr=$WellAddr",
        "--tls-ca-dir=$TLSCADir",
        "--beacon",
        "--beacon-ssid=$BeaconSSID"
    )
    if ($Interactive) {
        $argsList += "--interactive"
    }
    Push-Location $RepoRoot
    try {
        & go @argsList
    } finally {
        Pop-Location
    }
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight Toolchain & Repository Anchor Verification )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.room_well_host
# Mission:       Confirm Go toolchain and FaithNet go.mod anchor before boot.
# ==============================================================================

function Test-RoomWellPreflight {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "[PREFLIGHT FAIL] Go toolchain not found in PATH."
    }
    $goMod = Join-Path $RepoRoot "go.mod"
    if (-not (Test-Path $goMod)) {
        throw "[PREFLIGHT FAIL] FaithNet go.mod anchor missing at $RepoRoot."
    }
    Write-Host "[PREFLIGHT PASS] Windows Field Kit 01 Room & Well anchor verified ($RepoRoot)."
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Lifecycle Orchestration Entrypoint )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.room_well_host
# Mission:       Run pre-flight audit, show banner, and launch unless -AuditOnly.
# ==============================================================================

Test-RoomWellPreflight
$hostIP = Get-HostLANIPv4
Show-VisitorBanner -HostIP $hostIP
if ($AuditOnly) {
    Write-Host "[AUDIT ONLY] Parity pre-flight completed cleanly; skipping live server launch."
    exit 0
}
Invoke-RoomWellNode

# End of tests/playtester/field_kits/01_two_device_room_well/windows/run_room_well_host.ps1
