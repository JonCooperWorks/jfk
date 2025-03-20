# JFK Files Downloader

A respectful and efficient downloader for the JFK Files released by the National Archives (NARA). This tool is designed to download PDF files from NARA's website while being mindful of their server resources.

## Overview

This tool downloads PDF files from the National Archives' JFK Records collection, specifically targeting the 2025 release documents. It features concurrent downloads with rate limiting to prevent overwhelming the server.

## Prerequisites

- Go 1.23 or higher
- The `github.com/PuerkitoBio/goquery` package

## Installation

1. Clone the repository:
```bash
git clone https://github.com/joncooperworks/jfk.git
```

1.  Build the program:
```bash
go get ./...

go build
```

## Usage

1. First, download the HTML file containing the links from NARA's website and save it as `jfk-release-2025.html`

2. Run the downloader:
```bash
./downloader
```

### Available Flags

- `-file`: Path to the local HTML file to parse (default: "jfk-release-2025.html")
- `-base`: Base URL for resolving relative links (default: "https://www.archives.gov/research/jfk/release-2025")
- `-out`: Output directory for downloaded files (default: "pdfs")
- `-c`: Number of concurrent downloads (default: 5)
- `-ua`: User Agent string for HTTP requests

### Example

To download with 3 concurrent downloads to a custom directory:
```bash
./downloader -file jfk-release-2025.html -c 3 -out downloaded_pdfs
```

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
