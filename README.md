# JFK Files Downloader

A respectful and efficient downloader for the JFK Files released by the National Archives (NARA). This tool is designed to download PDF files from NARA's website while being mindful of their server resources, and can optionally perform OCR on downloaded files.

## Overview

This tool downloads PDF files from the National Archives' JFK Records collection, specifically targeting the 2025 release documents. It features concurrent downloads with rate limiting to prevent overwhelming the server. It can also extract text from PDFs, using OCR for scanned documents.

## Prerequisites

- Go 1.23 or higher
- The following Go packages:
  - `github.com/PuerkitoBio/goquery`
  - `github.com/otiai10/gosseract/v2`
  - `github.com/gen2brain/go-fitz`
- Tesseract OCR and Leptonica (for OCR functionality)

### Installing Tesseract OCR

For macOS:
```bash
brew install tesseract
brew install leptonica
brew install pkg-config
brew install tesseract-lang  # Optional: for additional language support

# Add these to your ~/.zshrc or ~/.bash_profile:
export PKG_CONFIG_PATH=/opt/homebrew/opt/tesseract/lib/pkgconfig:/opt/homebrew/opt/leptonica/lib/pkgconfig
export CGO_CFLAGS="-I/opt/homebrew/include"
export CGO_LDFLAGS="-L/opt/homebrew/lib"
```

For Ubuntu/Debian:
```bash
sudo apt-get install tesseract-ocr
sudo apt-get install libtesseract-dev
```

For Windows:
- Download and install Tesseract from the official GitHub releases

## Installation

1. Clone the repository:
```bash
git clone https://github.com/joncooperworks/jfk.git
```

2. Build the program:
```bash
go get ./...
go build
```

## Usage

1. Run the downloader:
```bash
./downloader
```

### Available Flags

- `-file`: Path to the local HTML file to parse (default: "jfk-release-2025.html")
- `-base`: Base URL for resolving relative links (default: "https://www.archives.gov/research/jfk/release-2025")
- `-out`: Output directory for downloaded files (default: "pdfs")
- `-c`: Number of concurrent downloads (default: 5)
- `-ua`: User Agent string for HTTP requests
- `-text`: Enable text extraction and OCR processing (creates .txt files alongside PDFs)

### Example

To download with 3 concurrent downloads to a custom directory and perform OCR:
```bash
./downloader -file jfk-release-2025.html -c 3 -out downloaded_pdfs -text
```

## Text Extraction

When the `-text` flag is enabled, the tool will:
1. Attempt to extract text directly from PDFs that have embedded text
2. For scanned documents or PDFs without extractable text, it will automatically perform OCR
3. Create a .txt file alongside each PDF containing the extracted text

## Ethical Usage

This tool is designed with respect for NARA's servers in mind. Please:

- Keep concurrent downloads to a reasonable number (default is 5)
- Don't modify the code to bypass rate limits
- Consider running downloads during off-peak hours
- Be patient with the download process

## Legal Note

The JFK Records are public domain documents made available by the National Archives. However, please ensure you comply with NARA's terms of service and usage guidelines when downloading these files.

## Contributing

Feel free to submit issues and enhancement requests!
