package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gen2brain/go-fitz"
	pdf "github.com/ledongthuc/pdf"
	"github.com/otiai10/gosseract/v2"
)

func main() {
	// Define command-line flags.
	htmlFile := flag.String("file", "jfk-release-2025.html", "Path to the local HTML file to parse")
	baseURLStr := flag.String("base", "https://www.archives.gov/research/jfk/release-2025", "Base URL for resolving relative links")
	outDir := flag.String("out", "pdfs", "Output directory for downloaded files")
	textOutDir := flag.String("textout", "text", "Output directory for converted text files")
	concurrency := flag.Int("c", 5, "Number of concurrent downloads")
	userAgent := flag.String("ua", "JFK-Files-Downloader/1.0 (Thank you President Trump for releasing these files!)", "User Agent string for HTTP requests")
	convertText := flag.Bool("text", false, "Convert PDFs to text files")
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

	// Pre-check existing files and filter out already downloaded PDFs
	var filteredPDFLinks []string
	for _, pdfURL := range pdfLinks {
		parts := strings.Split(pdfURL, "/")
		fileName := parts[len(parts)-1]
		filePath := filepath.Join(*outDir, fileName)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			filteredPDFLinks = append(filteredPDFLinks, pdfURL)
		} else {
			fmt.Printf("File %s already exists, skipping download\n", filePath)
		}
	}

	fmt.Printf("Downloading %d new PDFs\n", len(filteredPDFLinks))

	// Use a buffered channel as a semaphore to limit concurrency.
	sem := make(chan struct{}, *concurrency)
	var wg sync.WaitGroup

	// Download each PDF concurrently.
	for _, pdfURL := range filteredPDFLinks {
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

	// After downloads complete, convert to text if requested
	if *convertText {
		fmt.Println("Converting PDFs to text...")
		files, err := os.ReadDir(*outDir)
		if err != nil {
			fmt.Printf("Error reading output directory: %v\n", err)
			os.Exit(1)
		}

		// Create the text output directory if it doesn't exist
		if err := os.MkdirAll(*textOutDir, os.ModePerm); err != nil {
			fmt.Printf("Error creating text output directory: %v\n", err)
			os.Exit(1)
		}

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".pdf") {
				pdfPath := filepath.Join(*outDir, file.Name())
				textPath := filepath.Join(*textOutDir, strings.TrimSuffix(file.Name(), ".pdf")+".txt")

				// Skip if text file already exists
				if _, err := os.Stat(textPath); err == nil {
					continue
				}

				fmt.Printf("Converting %s to text...\n", file.Name())
				if err := convertPDFToText(pdfPath, textPath); err != nil {
					fmt.Printf("Error converting %s: %v\n", file.Name(), err)
					continue
				}
			}
		}
		fmt.Println("Text conversion completed.")
	}

	fmt.Println("All operations completed.")
}

func convertPDFToText(pdfPath, textPath string) error {
	// First try to extract text directly
	f, r, err := pdf.Open(pdfPath)
	if err != nil {
		return fmt.Errorf("error opening PDF: %v", err)
	}
	defer f.Close()

	textFile, err := os.Create(textPath)
	if err != nil {
		return fmt.Errorf("error creating text file: %v", err)
	}
	defer textFile.Close()

	totalPage := r.NumPage()
	hasText := false

	// Try direct text extraction first
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}

		if len(strings.TrimSpace(text)) > 0 {
			hasText = true
			fmt.Fprintf(textFile, "--- Page %d ---\n%s\n\n", pageIndex, text)
		}
	}

	// If no text was extracted, use OCR
	if !hasText {
		fmt.Printf("No text found in PDF, attempting OCR: %s\n", pdfPath)

		// Reset file position
		textFile.Seek(0, 0)

		// Open document with fitz (MuPDF)
		doc, err := fitz.New(pdfPath)
		if err != nil {
			return fmt.Errorf("error opening PDF for OCR: %v", err)
		}
		defer doc.Close()

		// Initialize Tesseract
		client := gosseract.NewClient()
		defer client.Close()

		// Process each page
		for n := 0; n < doc.NumPage(); n++ {
			img, err := doc.Image(n)
			if err != nil {
				fmt.Printf("Error getting page image %d: %v\n", n+1, err)
				continue
			}

			// Convert image to bytes
			var buf bytes.Buffer
			err = png.Encode(&buf, img)
			if err != nil {
				fmt.Printf("Error encoding image to bytes: %v\n", err)
				continue
			}

			// Set the image for OCR
			err = client.SetImageFromBytes(buf.Bytes())
			if err != nil {
				fmt.Printf("Error setting image for OCR: %v\n", err)
				continue
			}

			// Perform OCR
			text, err := client.Text()
			if err != nil {
				fmt.Printf("Error performing OCR on page %d: %v\n", n+1, err)
				continue
			}

			fmt.Fprintf(textFile, "--- Page %d (OCR) ---\n%s\n\n", n+1, text)
		}
	}

	return nil
}
