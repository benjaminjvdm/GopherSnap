# GopherSnap

GopherSnap is a high-performance, concurrent CLI image converter written in Go. It allows you to batch process images with efficiency, supporting modern formats like WebP and AVIF with zero external runtime C dependencies.

## Features

- **Batch Conversion**: Process entire directories of images at once.
- **Recursive Directory Mirroring**: Preserves input directory structure in the output.
- **Concurrent Processing**: Leverages Go's goroutines for fast, parallel execution.
- **Modern Formats**: Supports JPG, PNG, WebP, and AVIF.
- **Pure-Go & Zero CGO**: Compiles cleanly with `CGO_ENABLED=0` using WASM/pure-Go encoders for maximum cross-platform portability.
- **Standalone Windows Executable**: Embedded application icons and PE version manifests for Windows (`.exe`).
- **Interactive Progress**: Real-time feedback with a styled progress bar.

## Architecture Overview

GopherSnap is architected for zero-CGO portability and parallel performance:

- **Core Converter Pipeline**: Pure Go image decoding (`image/jpeg`, `image/png`, `image/gif`) combined with pure-Go and WASM-backed WebP (`github.com/gen2brain/webp`) and AVIF (`github.com/gen2brain/avif`) encoders via Wazero.
- **Concurrency Model**: Worker pool pattern (`internal/converter/batch.go`) distributing file transformation jobs across customizable goroutine workers.
- **Windows Embedding**: Windows application icon and PE metadata are compiled into native `rsrc_windows_*.syso` resources using `go-winres`.
- **Cross-Platform Utility**: Clean path sanitization and case-insensitive image extension matching (`.jpg`, `.jpeg`, `.png`, `.webp`, `.avif`, `.gif`).

## Installation

### Windows (.EXE) Standalone Quickstart

Download `gophersnap_windows_x64.zip` from the latest GitHub Release, extract `gophersnap.exe`, and run it directly in PowerShell or Command Prompt:

```powershell
# Extract gophersnap.exe and test version
.\gophersnap.exe version

# Convert images in PowerShell
.\gophersnap.exe convert -i C:\Images\Photos -o C:\Images\Output -f webp -q 85
```

### Via Go Install

Ensure Go is installed on your system:

```bash
go install github.com/benjaminjvdm/GopherSnap@latest
```

### Build From Source

Clone the repository and build using Makefile or PowerShell build automation:

```bash
git clone https://github.com/benjaminjvdm/GopherSnap.git
cd GopherSnap

# Linux / macOS (Unix)
make build

# Windows PowerShell
.\build.ps1
```

## Build Automation & PowerShell Scripts

### Windows PowerShell (`build.ps1`)

```powershell
# Default build: compiles standalone gophersnap.exe with stripped symbols
.\build.ps1

# Run full test suite with coverage report
.\build.ps1 -Test

# Generate Windows PE resources and package release ZIP
.\build.ps1 -Package
```

### Unix Makefile (`Makefile`)

```bash
# Build standalone binary
make build

# Run test suite
make test

# Generate Windows PE resources and release package
make release

# Clean binaries and build artifacts
make clean
```

## Usage

GopherSnap provides a simple `convert` command to handle your image processing needs.

### Basic Example

Convert all images in a folder to WebP. GopherSnap will automatically mirror the input directory's structure in the output directory:

```bash
gophersnap convert -i ./input-images -o ./output-images -f webp
```

### Advanced Usage

Convert a specific file to AVIF with custom quality and concurrency:

```bash
gophersnap convert -i photo.jpg -o ./optimized -f avif -q 75 -j 8
```

### Image Resizing

Resize images while maintaining aspect ratio:

```bash
# Resize to 800px width (height calculated automatically)
gophersnap convert -i ./images --width 800

# Resize to fit within 1024x1024
gophersnap convert -i ./images --width 1024 --height 1024
```

### Available Flags

- `-i, --input string`: Input file or directory (Required)
- `-o, --output string`: Output directory (Default: `./output`)
- `-f, --format string`: Output format: `jpg`, `png`, `webp`, `avif` (Default: `webp`)
- `-q, --quality int`: Image quality (0-100) (Default: `80`)
- `--max-size string`: Maximum file size (e.g., `500kb`, `1mb`)
- `--width int`: Target width (maintaining aspect ratio)
- `--height int`: Target height (maintaining aspect ratio)
- `-j, --jobs int`: Number of concurrent jobs (Default: `4`)
- `--overwrite`: Overwrite existing files if they exist in the output directory

## Version & Release Pipeline

GopherSnap features automated cross-platform releases powered by GoReleaser and GitHub Actions.

### Version Command & Linker Injection

Build-time metadata can be injected via ldflags:

```bash
go build -ldflags="-s -w \
  -X 'github.com/benjaminjvdm/GopherSnap/cmd.Version=v0.1.0' \
  -X 'github.com/benjaminjvdm/GopherSnap/cmd.Commit=1a2b3c4' \
  -X 'github.com/benjaminjvdm/GopherSnap/cmd.Date=2026-03-30'" \
  -o gophersnap main.go
```

Check installed version details:

```bash
gophersnap version
# or
gophersnap --version
```

### Release Process & GoReleaser Configuration

To trigger an automated release:
1. Commit all changes and create a semantic version tag: `git tag -a v0.1.0 -m "Release v0.1.0"`.
2. Push the tag: `git push origin v0.1.0`.
3. GitHub Actions (`.github/workflows/release.yml`) executes `.goreleaser.yaml` to build cross-platform binaries:
   - **Windows**: `windows/amd64`, `windows/arm64` (.zip with `gophersnap.exe`, `README.md`, `LICENSE`)
   - **Linux**: `linux/amd64`, `linux/arm64` (.tar.gz)
   - **macOS**: `darwin/amd64`, `darwin/arm64` (Universal binary .tar.gz)
   - **Checksums**: SHA-256 (`checksums.txt`)
