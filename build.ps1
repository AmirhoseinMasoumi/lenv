# Lenv Build Script for Windows
# Usage: .\build.ps1 [command]

param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "build-all", "test", "clean", "install", "dev", "help")]
    [string]$Command = "build"
)

$BinaryName = "lenv"
$BuildDir = "build"

# Get version from git
try {
    $Version = git describe --tags --always --dirty 2>$null
    if (-not $Version) { $Version = "dev" }
} catch {
    $Version = "dev"
}

$BuildTime = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$LdFlags = "-ldflags `"-X main.Version=$Version -X main.BuildTime=$BuildTime`""

function Write-Header {
    param([string]$Text)
    Write-Host ""
    Write-Host "=== $Text ===" -ForegroundColor Cyan
}

function Build {
    Write-Header "Building lenv"
    if (-not (Test-Path $BuildDir)) {
        New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
    }
    
    $cmd = "go build $LdFlags -o $BuildDir/$BinaryName.exe ."
    Write-Host "Running: $cmd" -ForegroundColor Gray
    Invoke-Expression $cmd
    
    if ($LASTEXITCODE -eq 0) {
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
    
    # Windows
    Write-Host "Building Windows..." -ForegroundColor Yellow
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build $LdFlags -o "$BuildDir/$BinaryName.exe" .
    
    # Linux
    Write-Host "Building Linux..." -ForegroundColor Yellow
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    go build $LdFlags -o "$BuildDir/$BinaryName-linux" .
    
    # macOS
    Write-Host "Building macOS..." -ForegroundColor Yellow
    $env:GOOS = "darwin"
    $env:GOARCH = "amd64"
    go build $LdFlags -o "$BuildDir/$BinaryName-darwin" .
    
    # Reset environment
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    
    Write-Host "All builds complete!" -ForegroundColor Green
    Get-ChildItem $BuildDir | Format-Table Name, Length
}

function Test {
    Write-Header "Running tests"
    go test -v ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Tests failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host "All tests passed!" -ForegroundColor Green
}

function Clean {
    Write-Header "Cleaning build artifacts"
    if (Test-Path $BuildDir) {
        Remove-Item -Recurse -Force $BuildDir
        Write-Host "Removed $BuildDir/" -ForegroundColor Yellow
    }
    if (Test-Path "$BinaryName.exe") {
        Remove-Item -Force "$BinaryName.exe"
        Write-Host "Removed $BinaryName.exe" -ForegroundColor Yellow
    }
    if (Test-Path $BinaryName) {
        Remove-Item -Force $BinaryName
        Write-Host "Removed $BinaryName" -ForegroundColor Yellow
    }
    Write-Host "Clean complete!" -ForegroundColor Green
}

function Install {
    Write-Header "Installing to GOPATH"
    $gopath = go env GOPATH
    $installPath = Join-Path $gopath "bin"
    
    go build $LdFlags -o "$installPath/$BinaryName.exe" .
    
    if ($LASTEXITCODE -eq 0) {
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
    
    go build -race $LdFlags -o "$BuildDir/$BinaryName.exe" .
    
    if ($LASTEXITCODE -eq 0) {
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
    Write-Host "Usage: .\build.ps1 [command]" -ForegroundColor White
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

# Execute command
switch ($Command) {
    "build"     { Build }
    "build-all" { BuildAll }
    "test"      { Test }
    "clean"     { Clean }
    "install"   { Install }
    "dev"       { Dev }
    "help"      { ShowHelp }
}
