package main

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// App struct
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

var client = &http.Client{
	Timeout: 10 * time.Second,
}

type Response struct {
	Response  string
	Status    string
	Size      string
	TotalTime float64
	Headers   http.Header
}

func (a *App) SendRequest(URL string, method string) Response {
	start := time.Now()

	req, err := http.NewRequest(method, URL, nil)
	if err != nil {
		return logError("Error creating the request:", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return logError("Error making the request:", err)
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		// Handle gzip encoding
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return logError("Error creating gzip reader:", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Read and process the response in chunks
	const bufferSize = 8192
	buffer := make([]byte, bufferSize)
	var bodyContent strings.Builder
	totalSize := 0

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			bodyContent.Write(buffer[:n])
			totalSize += n
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return logError("Error reading the response body:", err)
		}
	}

	contentType := resp.Header.Get("Content-Type")
	size := formatSize(totalSize)
	output := handleResponseContent(contentType, []byte(bodyContent.String()))
	roundedTime := math.Round(time.Since(start).Seconds()*1000*100) / 100

	return Response{output, resp.Status, size, roundedTime, resp.Header}
}

func logError(message string, err error) Response {
	fmt.Println(message, err)
	return Response{}
}

func formatSize(sizeBytes int) string {
	switch {
	case sizeBytes >= 1024*1024:
		return fmt.Sprintf("%.2f MB", float64(sizeBytes)/(1024*1024))
	case sizeBytes >= 1024:
		return fmt.Sprintf("%.2f KB", float64(sizeBytes)/1024)
	default:
		return fmt.Sprintf("%d bytes", sizeBytes)
	}
}

func handleResponseContent(contentType string, body []byte) string {
	switch {
	case strings.Contains(contentType, "application/json"):
		var response interface{}
		if err := json.Unmarshal(body, &response); err != nil {
			fmt.Println("Error parsing JSON:", err)
			return ""
		}
		prettyJSON, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			fmt.Println("Error formatting JSON:", err)
			return ""
		}
		return html.UnescapeString(string(prettyJSON))

	case strings.Contains(contentType, "text/html"):
		return string(body)

	default:
		return string(body)
	}
}
