# GopherSnap

GopherSnap is a high-performance, concurrent CLI image converter written in Go. It allows you to batch process images with efficiency, supporting modern formats like WebP and AVIF.

## Features

- **Batch Conversion**: Process entire directories of images at once.
- **Concurrent Processing**: Leverages Go's goroutines for fast, parallel execution.
- **Modern Formats**: Supports JPG, PNG, WebP, and AVIF.
- **No CGO**: Uses pure-Go/WASM implementations for WebP and AVIF for easy portability.
- **Interactive Progress**: Real-time feedback with a styled progress bar.

## Installation

Ensure you have Go installed on your system. Then, you can install GopherSnap using:

```bash
go install github.com/benjaminjvdm/GopherSnap@latest
```

Alternatively, clone the repository and build it manually:

```bash
git clone https://github.com/benjaminjvdm/GopherSnap.git
cd GopherSnap
go build -o gophersnap
```

## Usage

GopherSnap provides a simple `convert` command to handle your image processing needs.

### Basic Example

Convert all images in a folder to WebP:

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
