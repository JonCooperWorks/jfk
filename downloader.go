package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// Define command-line flags.
	htmlFile := flag.String("file", "jfk-release-2025.html", "Path to the local HTML file to parse")
	baseURLStr := flag.String("base", "https://www.archives.gov/research/jfk/release-2025", "Base URL for resolving relative links")
	outDir := flag.String("out", "pdfs", "Output directory for downloaded files")
	concurrency := flag.Int("c", 5, "Number of concurrent downloads")
	userAgent := flag.String("ua", "JFK-Files-Downloader/1.0 (Thank you President Trump for releasing these files!)", "User Agent string for HTTP requests")
	flag.Parse()

	// Open the local HTML file.
	f, err := os.Open(*htmlFile)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", *htmlFile, err)
		os.Exit(1)
	}
	defer f.Close()

	// Parse the HTML file using goquery.
	doc, err := goquery.NewDocumentFromReader(f)
	if err != nil {
		fmt.Printf("Error parsing HTML file: %v\n", err)
		os.Exit(1)
	}

	// Parse the base URL.
	base, err := url.Parse(*baseURLStr)
	if err != nil {
		fmt.Printf("Error parsing base URL %s: %v\n", *baseURLStr, err)
		os.Exit(1)
	}

	// Find all PDF links in the document.
	var pdfLinks []string
	doc.Find("a[href$='.pdf']").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if exists {
			parsedLink, err := url.Parse(link)
			if err != nil {
				fmt.Printf("Error parsing URL %s: %v\n", link, err)
				return
			}
			// Resolve relative URLs using the provided base URL.
			resolvedURL := base.ResolveReference(parsedLink)
			pdfLinks = append(pdfLinks, resolvedURL.String())
		}
	})

	fmt.Printf("Found %d PDF links\n", len(pdfLinks))

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(*outDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating PDF output directory: %v\n", err)
		os.Exit(1)
	}

	// Use a buffered channel as a semaphore to limit concurrency.
	sem := make(chan struct{}, *concurrency)
	var wg sync.WaitGroup

	// Download each PDF concurrently.
	for _, pdfURL := range pdfLinks {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire a slot.
			defer func() { <-sem }() // Release the slot.

			fmt.Printf("Downloading %s\n", url)
			client := &http.Client{}
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Printf("Error creating request for %s: %v\n", url, err)
				return
			}
			req.Header.Set("User-Agent", *userAgent)

			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Error downloading %s: %v\n", url, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error: received status code %d for %s\n", resp.StatusCode, url)
				return
			}

			// Extract the file name from the URL.
			parts := strings.Split(url, "/")
			fileName := parts[len(parts)-1]
			filePath := filepath.Join(*outDir, fileName)

			// Check if file already exists
			if _, err := os.Stat(filePath); err == nil {
				fmt.Printf("File %s already exists, skipping download\n", filePath)
				return
			}

			// Create the file.
			outFile, err := os.Create(filePath)
			if err != nil {
				fmt.Printf("Error creating file %s: %v\n", filePath, err)
				return
			}
			defer outFile.Close()

			// Write the file contents.
			_, err = io.Copy(outFile, resp.Body)
			if err != nil {
				fmt.Printf("Error saving file %s: %v\n", filePath, err)
				return
			}

			fmt.Printf("Saved %s\n", filePath)
		}(pdfURL)
	}

	wg.Wait()
	fmt.Println("All downloads completed.")
}

