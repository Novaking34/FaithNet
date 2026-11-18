# ==============================================================================
# BLOCK 1: METADATA BLOCK
# ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
# ==============================================================================
# @project       FaithNet
# @file          tests/playtester/field_kits/01_two_device_room_well/windows/run_wifi_ap_router.ps1
# @id            playtest.field_kit.01.windows.wifi_ap_router
# @layer         Layer 4: Proving Ground (Field Kit 01 — Windows Mobile Hotspot)
# @archetype     field_script | operational
# @version       0.1.0-dev
# @interface_rev Rev 5
# @date          2026-11-17
# @updated       2026-11-17
# @author        CreativeWorkzStudio LLC / ProfessorSeanEX
# @repository    https://github.com/CreativeWorkzStudio/FaithNet
# @system        FaithNet Simultaneous STA+AP Wi-Fi Router Launcher (Windows 10/11)
# @role          Configure Windows Mobile Hotspot + Private Firewall Rules
# @maturity      Rev 5 Baseline | Wave 14 5-Block Cross-Platform Parity
# @chassis       winrt-native | Uses Windows native Mobile Hotspot & ICS safely
# @scale         Simultaneous STA+AP via WinRT NetworkOperatorTetheringManager
# @seams         in: WinRT TetheringManager, NetFirewallRule | exports: FaithNet AP
# @invariants    1. Universal 5-Block Lifecycle Protocol (Rev 5).
#                2. Zero-brick: uses official Windows Mobile Hotspot API instead
#                   of deprecated `netsh wlan hostednetwork`.
#                3. Supports `-AuditOnly` for safe cross-platform pwsh parity tests.
#
# PURPOSE & ARCHITECTURAL ROLE:
#   Windows 10/11 counterpart to Linux `run_wifi_ap_router.sh`. Opens minimal
#   inbound firewall rules for FaithNet (TCP 8077 Well + UDP 7777 LAN) on the
#   Private profile and activates Windows native Mobile Hotspot (Wi-Fi Direct
#   SoftAP) broadcasting `FaithNet-HomeRoom` while keeping upstream Wi-Fi live.
#
#   ### For Architects, Operators & Non-Programmers:
#   Turns a Windows 10/11 laptop into a FaithNet Wi-Fi router using Windows'
#   built-in safe Mobile Hotspot engine, without risky third-party drivers.
#
#   ### For Systems Programmers & Engine Architects:
#   Calls WinRT `NetworkOperatorTetheringManager` on Windows hosts and degrades
#   cleanly with an explicit diagnostic summary on non-Windows `pwsh` hosts.
# ==============================================================================
[CmdletBinding()]
param(
    [string]$SSID = "FaithNet-HomeRoom",
    [string]$Pass = "FaithNet2026",
    [switch]$AuditOnly
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# ==============================================================================
# BLOCK 2: OPENING BLOCK
# ( Operational Constants & Firewall Rule Identifiers )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.wifi_ap_router
# Mission:       Pin SSID, WPA2 passphrase bounds, and firewall rule names.
# ==============================================================================

$WellRuleName = "FaithNet-TheWell-TCP-8077"
$UDPRuleName  = "FaithNet-Rung3-UDP-7777"

# ==============================================================================
# BLOCK 3: BODY BLOCK
# ( Windows Firewall Allowance & Native Mobile Hotspot Activation )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.wifi_ap_router
# Mission:       Allow Private-profile FaithNet ports and start WinRT SoftAP.
# ==============================================================================

function Enable-FaithNetWindowsFirewallRules {
    if (-not (Get-Command New-NetFirewallRule -ErrorAction SilentlyContinue)) {
        Write-Host "[INFO] NetSecurity module unavailable on this host; skipping Windows Firewall step."
        return
    }
    if (-not (Get-NetFirewallRule -DisplayName $WellRuleName -ErrorAction SilentlyContinue)) {
        New-NetFirewallRule -DisplayName $WellRuleName -Direction Inbound -Protocol TCP -LocalPort 8077 -Action Allow -Profile Private | Out-Null
    }
    if (-not (Get-NetFirewallRule -DisplayName $UDPRuleName -ErrorAction SilentlyContinue)) {
        New-NetFirewallRule -DisplayName $UDPRuleName -Direction Inbound -Protocol UDP -LocalPort 7777 -Action Allow -Profile Private | Out-Null
    }
    Write-Host "[OK] Windows Defender Firewall rules active for TCP 8077 & UDP 7777 (Private profile)."
}

function Start-WindowsMobileHotspot {
    Write-Host "=== FAITHNET WI-FI AP ROUTER (WINDOWS 10/11 MOBILE HOTSPOT) ==="
    Write-Host "Target SSID : $SSID"
    Write-Host "WPA2 Pass   : $Pass"
    Write-Host "Well Portal : https://<Hotspot-Gateway-IP>:8077/ (typically 192.168.137.1:8077)"
    Write-Host ""
    if (-not $IsWindows -and $PSVersionTable.PSEdition -eq "Core") {
        Write-Host "[INFO] Running under non-Windows PowerShell ($($PSVersionTable.OS))."
        Write-Host "[INFO] On Windows 10/11, this activates WinRT NetworkOperatorTetheringManager cleanly."
        return
    }

    Add-Type -AssemblyName System.Runtime.WindowsRuntime
    $profile = [Windows.Networking.Connectivity.NetworkInformation,Windows.Networking.Connectivity,ContentType=WindowsRuntime]::GetInternetConnectionProfile()
    if ($null -eq $profile) {
        Write-Host "[WARN] No upstream internet profile detected; enable Mobile Hotspot manually in Settings -> Network & internet -> Mobile hotspot."
        return
    }
    $mgr = [Windows.Networking.NetworkOperators.NetworkOperatorTetheringManager,Windows.Networking.NetworkOperators,ContentType=WindowsRuntime]::CreateFromConnectionProfile($profile)
    $cfg = $mgr.GetCurrentAccessPointConfiguration()
    $cfg.Ssid = $SSID
    $cfg.Passphrase = $Pass
    $null = $mgr.ConfigureAccessPointAsync($cfg)
    $null = $mgr.StartTetheringAsync()
    Write-Host "[LIVE] Windows Mobile Hotspot '$SSID' start requested cleanly!"
}

# ==============================================================================
# BLOCK 4: AUDIT PROCEDURES BLOCK
# ( Pre-Flight SSID & WPA2 Passphrase Length Validation )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.wifi_ap_router
# Mission:       Enforce IEEE 802.11 SSID (1..32B) and WPA2 (8..63B) bounds.
# ==============================================================================

function Test-WindowsAPPreflight {
    if ($SSID.Length -lt 1 -or $SSID.Length -gt 32) {
        throw "[PREFLIGHT FAIL] SSID must be 1..32 characters (IEEE 802.11 bound)."
    }
    if ($Pass.Length -lt 8 -or $Pass.Length -gt 63) {
        throw "[PREFLIGHT FAIL] WPA2 passphrase must be 8..63 characters."
    }
    Write-Host "[PREFLIGHT PASS] Windows Mobile Hotspot parameters verified (SSID='$SSID')."
}

# ==============================================================================
# BLOCK 5: CLOSING BLOCK
# ( Lifecycle Entrypoint )
# ==============================================================================
# Engine:        FaithNet Adaptive Networking Substrate
# Stratum:       Layer 4: Proving Ground (Field Kit 01 — Windows)
# Module:        playtest.field_kit.01.windows.wifi_ap_router
# Mission:       Run pre-flight, configure firewall, and start Mobile Hotspot.
# ==============================================================================

Test-WindowsAPPreflight
if ($AuditOnly) {
    Write-Host "[AUDIT ONLY] Windows AP script pre-flight verified cleanly."
    exit 0
}
Enable-FaithNetWindowsFirewallRules
Start-WindowsMobileHotspot

# End of tests/playtester/field_kits/01_two_device_room_well/windows/run_wifi_ap_router.ps1
