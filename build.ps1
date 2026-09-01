[CmdletBinding()]
param(
    [switch]$Test,
    [switch]$Package
)

$ErrorActionPreference = "Stop"

if ($Test) {
    Write-Host "Running full test suite..." -ForegroundColor Cyan
    go test -v -cover ./...
    exit $LASTEXITCODE
}

if ($Package) {
    Write-Host "Generating Windows PE resources..." -ForegroundColor Cyan
    go run github.com/tc-hib/go-winres@latest make --arch amd64,arm64,386

    Write-Host "Building gophersnap.exe..." -ForegroundColor Cyan
    $env:CGO_ENABLED = "0"
    $env:GOOS = "windows"
    $env:GOARCH = "amd64"
    go build -ldflags="-s -w" -o gophersnap.exe main.go

    Write-Host "Packaging gophersnap_windows_x64.zip..." -ForegroundColor Cyan
    if (Test-Path "gophersnap_windows_x64.zip") {
        Remove-Item "gophersnap_windows_x64.zip" -Force
    }
    Compress-Archive -Path "gophersnap.exe", "README.md", "LICENSE" -DestinationPath "gophersnap_windows_x64.zip"
    Write-Host "Successfully packaged gophersnap_windows_x64.zip" -ForegroundColor Green
    exit 0
}

Write-Host "Building standalone gophersnap.exe..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
go build -ldflags="-s -w" -o gophersnap.exe main.go
Write-Host "Build complete: gophersnap.exe" -ForegroundColor Green
