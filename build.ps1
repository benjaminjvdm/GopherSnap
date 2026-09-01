[CmdletBinding()]
param(
    [switch]$Test,
    [switch]$Package
)

$ErrorActionPreference = "Stop"

$BinDir = "binaries"
if (-not (Test-Path $BinDir)) {
    New-Item -ItemType Directory -Path $BinDir | Out-Null
}

if ($Test) {
    Write-Host "Running full test suite..." -ForegroundColor Cyan
    go test -v -cover ./...
    exit $LASTEXITCODE
}

if ($Package) {
    Write-Host "Generating Windows PE resources..." -ForegroundColor Cyan
    go run github.com/tc-hib/go-winres@latest make --arch amd64,arm64,386

    Write-Host "Building binaries/gophersnap.exe..." -ForegroundColor Cyan
    $env:CGO_ENABLED = "0"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -ldflags="-s -w" -o "$BinDir/gophersnap.exe" main.go

    Write-Host "Packaging binaries/gophersnap_windows_x64.zip..." -ForegroundColor Cyan
    $ZipPath = "$BinDir/gophersnap_windows_x64.zip"
    if (Test-Path $ZipPath) {
        Remove-Item $ZipPath -Force
    }
    Compress-Archive -Path "$BinDir/gophersnap.exe", "README.md", "LICENSE" -DestinationPath $ZipPath
    Write-Host "Successfully packaged $ZipPath" -ForegroundColor Green
    exit 0
}

Write-Host "Building standalone binaries/gophersnap.exe..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
go build -ldflags="-s -w" -o "$BinDir/gophersnap.exe" main.go
Write-Host "Build complete: $BinDir/gophersnap.exe" -ForegroundColor Green
