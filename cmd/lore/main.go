package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	// Define command line flags
	searchQuery := flag.String("q", "", "Search query")
	outputFile := flag.String("o", "output.html", "Output file name")
	flag.Parse()

	// Create URL with query parameters
	baseURL := "https://lore.kernel.org/lkml/"
	params := url.Values{}
	params.Add("q", *searchQuery)
	params.Add("x", "m")
	fullURL := baseURL + "?" + params.Encode()

	// Create a new HTTP client
	client := &http.Client{}

	// Create form data
	formData := "x=full+threads"

	// Create a new POST request
	req, err := http.NewRequest("POST", fullURL, strings.NewReader(formData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Set headers exactly as in curl
	req.Header = http.Header{
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language":           {"en-US,en;q=0.5"},
		"Cache-Control":             {"max-age=0"},
		"Content-Type":              {"application/x-www-form-urlencoded"},
		"Origin":                    {"https://lore.kernel.org"},
		"Priority":                  {"u=0, i"},
		"Referer":                   {"https://lore.kernel.org/lkml/?q=io_uring"},
		"Sec-Ch-Ua":                 {`"Brave";v="135", "Not-A.Brand";v="8", "Chromium";v="135"`},
		"Sec-Ch-Ua-Mobile":          {"?0"},
		"Sec-Ch-Ua-Platform":        {"Linux"},
		"Sec-Fetch-Dest":            {"document"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-Site":            {"same-origin"},
		"Sec-Fetch-User":            {"?1"},
		"Sec-Gpc":                   {"1"},
		"Upgrade-Insecure-Requests": {"1"},
		"User-Agent":                {"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"},
	}

	// Add cookie exactly as in curl
	cookie := &http.Cookie{
		Name:  "within.website-x-cmd-anubis-auth",
		Value: "eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJjaGFsbGVuZ2UiOiI1YmNkNWFkZWM4YzY0YjVmZWJkNzAwMTMwMWJhZTg5NTRiNGI3MGNhNmYwM2IzMjk3M2QyYjYzMDhmNzM0YjUwIiwiZXhwIjoxNzQ3MTU4NTU1LCJpYXQiOjE3NDY1NTM3NTUsIm5iZiI6MTc0NjU1MzY5NSwibm9uY2UiOjUyMDc4LCJyZXNwb25zZSI6IjAwMDBiZDc0OTI2YzcxOGRjNWVlNGEzMDBhZDVkOWNiZDg4NjkxZTQ0ZTAzODdhOTIxM2ZlNzg2ZWVhZjRmYjgifQ.jnvAPbL7GKqXr47ZXTc9ZkoZr4evUs690iL3KWcOstP4aw48xHVmJ-nCzeHLk-Hx_j3YucTl9nLXH3h2rtz1DQ",
	}
	req.AddCookie(cookie)

	// Start timing
	startTime := time.Now()
	lastUpdate := startTime
	var lastBytes int64

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Create output file
	file, err := os.Create(*outputFile)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Create a multi-writer to write to both file and count bytes
	var totalBytes int64
	writer := io.MultiWriter(file, &byteCounter{&totalBytes})

	// Create a ticker for updates
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Start a goroutine to print updates
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				elapsed := now.Sub(lastUpdate).Seconds()
				bytesSinceLast := totalBytes - lastBytes
				speed := float64(bytesSinceLast) / 1024 / elapsed // KB/s

				// Clear the current line and print new stats
				fmt.Printf("\rDownloading... Speed: %.2f KB/s, Total: %.2f MB",
					speed, float64(totalBytes)/1024/1024)

				lastUpdate = now
				lastBytes = totalBytes
			case <-done:
				return
			}
		}
	}()

	// Handle gzip content
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Type") == "application/gzip" {
		// Create a gzip reader directly from the response body
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			fmt.Printf("Error creating gzip reader: %v\n", err)
			os.Exit(1)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Copy response body to file and count bytes
	_, err = io.Copy(writer, reader)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		os.Exit(1)
	}

	// Stop the update goroutine
	done <- true

	// Calculate final duration
	duration := time.Since(startTime)

	// Print final statistics
	fmt.Printf("\n\nFinal Statistics:\n")
	fmt.Printf("------------------\n")
	fmt.Printf("Total Size: %.2f MB\n", float64(totalBytes)/1024/1024)
	fmt.Printf("Total Time: %v\n", duration)
	fmt.Printf("Average Speed: %.2f KB/s\n", float64(totalBytes)/1024/duration.Seconds())
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Response saved to %s\n", *outputFile)
}

// byteCounter implements io.Writer to count bytes
type byteCounter struct {
	total *int64
}

func (b *byteCounter) Write(p []byte) (int, error) {
	*b.total += int64(len(p))
	return len(p), nil
}
