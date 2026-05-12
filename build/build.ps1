# Lenv Build Script for Windows
# Usage: .\build\build.ps1 [command]
#
# This script lives in build/ and writes output here. It runs `go build`
# against the repo root (its parent directory), so it works regardless of
# the caller's current working directory.

param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "build-all", "test", "clean", "install", "dev", "help")]
    [string]$Command = "build"
)

$BinaryName = "lenv"
$BuildDir = $PSScriptRoot
$RepoRoot = Split-Path -Parent $PSScriptRoot

# Get version from git
Push-Location $RepoRoot
try {
    $Version = git describe --tags --always --dirty 2>$null
    if (-not $Version) { $Version = "dev" }
} catch {
    $Version = "dev"
}
Pop-Location

$BuildTime = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$LdFlags = "-ldflags `"-X main.Version=$Version -X main.BuildTime=$BuildTime`""

function Write-Header {
    param([string]$Text)
    Write-Host ""
    Write-Host "=== $Text ===" -ForegroundColor Cyan
}

function Invoke-GoBuild {
    param([string]$Output, [string[]]$ExtraArgs = @())
    Push-Location $RepoRoot
    try {
        $cmd = "go build $($ExtraArgs -join ' ') $LdFlags -o `"$Output`" ."
        Write-Host "Running: $cmd" -ForegroundColor Gray
        Invoke-Expression $cmd
        return $LASTEXITCODE
    } finally {
        Pop-Location
    }
}

function Build {
    Write-Header "Building lenv"
    if (-not (Test-Path $BuildDir)) {
        New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
    }
    $rc = Invoke-GoBuild -Output (Join-Path $BuildDir "$BinaryName.exe")
    if ($rc -eq 0) {
        Write-Host "Build successful: $BuildDir/$BinaryName.exe" -ForegroundColor Green
    } else {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
}

function BuildAll {
    Write-Header "Building for all platforms"
    if (-not (Test-Path $BuildDir)) {
        New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
    }

    Write-Host "Building Windows..." -ForegroundColor Yellow
    $env:GOOS = "windows"; $env:GOARCH = "amd64"
    Invoke-GoBuild -Output (Join-Path $BuildDir "$BinaryName.exe") | Out-Null

    Write-Host "Building Linux..." -ForegroundColor Yellow
    $env:GOOS = "linux"; $env:GOARCH = "amd64"
    Invoke-GoBuild -Output (Join-Path $BuildDir "$BinaryName-linux") | Out-Null

    Write-Host "Building macOS..." -ForegroundColor Yellow
    $env:GOOS = "darwin"; $env:GOARCH = "amd64"
    Invoke-GoBuild -Output (Join-Path $BuildDir "$BinaryName-darwin") | Out-Null

    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue

    Write-Host "All builds complete!" -ForegroundColor Green
    Get-ChildItem $BuildDir | Format-Table Name, Length
}

function Test {
    Write-Header "Running tests"
    Push-Location $RepoRoot
    try {
        go test -v ./...
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Tests failed!" -ForegroundColor Red
            exit 1
        }
    } finally {
        Pop-Location
    }
    Write-Host "All tests passed!" -ForegroundColor Green
}

function Clean {
    Write-Header "Cleaning build artifacts"
    Get-ChildItem -Path $BuildDir -File -ErrorAction SilentlyContinue | Where-Object {
        $_.Name -like "$BinaryName*" -or $_.Name -like "coverage.*"
    } | ForEach-Object {
        Remove-Item -Force $_.FullName
        Write-Host "Removed $($_.FullName)" -ForegroundColor Yellow
    }
    foreach ($stray in @((Join-Path $RepoRoot "$BinaryName.exe"), (Join-Path $RepoRoot $BinaryName))) {
        if (Test-Path $stray) {
            Remove-Item -Force $stray
            Write-Host "Removed $stray" -ForegroundColor Yellow
        }
    }
    Write-Host "Clean complete!" -ForegroundColor Green
}

function Install {
    Write-Header "Installing to GOPATH"
    $gopath = (& go env GOPATH).Trim()
    $installPath = Join-Path $gopath "bin"
    $rc = Invoke-GoBuild -Output (Join-Path $installPath "$BinaryName.exe")
    if ($rc -eq 0) {
        Write-Host "Installed to: $installPath/$BinaryName.exe" -ForegroundColor Green
    } else {
        Write-Host "Install failed!" -ForegroundColor Red
        exit 1
    }
}

function Dev {
    Write-Header "Building with race detector"
    if (-not (Test-Path $BuildDir)) {
        New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
    }
    $rc = Invoke-GoBuild -Output (Join-Path $BuildDir "$BinaryName.exe") -ExtraArgs @("-race")
    if ($rc -eq 0) {
        Write-Host "Dev build successful: $BuildDir/$BinaryName.exe" -ForegroundColor Green
    } else {
        Write-Host "Build failed!" -ForegroundColor Red
        exit 1
    }
}

function ShowHelp {
    Write-Host ""
    Write-Host "Lenv Build Script" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Usage: .\build\build.ps1 [command]" -ForegroundColor White
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor Yellow
    Write-Host "  build      Build binary for Windows (default)"
    Write-Host "  build-all  Build for all platforms (Windows, Linux, macOS)"
    Write-Host "  test       Run tests"
    Write-Host "  clean      Remove build artifacts"
    Write-Host "  install    Install to GOPATH/bin"
    Write-Host "  dev        Build with race detector"
    Write-Host "  help       Show this help"
    Write-Host ""
}

switch ($Command) {
    "build"     { Build }
    "build-all" { BuildAll }
    "test"      { Test }
    "clean"     { Clean }
    "install"   { Install }
    "dev"       { Dev }
    "help"      { ShowHelp }
}
