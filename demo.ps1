$ErrorActionPreference = "Stop"

Set-Location "D:\Github_Workstation\GitProjects\Lenv"

function Invoke-DemoStep {
    param(
        [Parameter(Mandatory = $true)][string]$Command,
        [int]$PauseSeconds = 2
    )

    Write-Host ""
    Write-Host "PS $((Get-Location).Path)> $Command"
    Invoke-Expression $Command
    Start-Sleep -Seconds $PauseSeconds
}

Invoke-DemoStep -Command ".\lenv.exe --help" -PauseSeconds 2
Invoke-DemoStep -Command ".\lenv.exe release-notes --version v1.1.0" -PauseSeconds 2
Invoke-DemoStep -Command ".\lenv.exe profile list" -PauseSeconds 2
Invoke-DemoStep -Command ".\lenv.exe runtime status" -PauseSeconds 2
Invoke-DemoStep -Command ".\lenv.exe provenance" -PauseSeconds 2
