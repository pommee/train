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

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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

func (a *App) SendRequest(URL, method string) Response {
	start := time.Now()

	req, err := http.NewRequest(method, URL, nil)
	if err != nil {
		return a.logError("Error creating the request:", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return a.logError("Error making the request:", err)
	}
	defer resp.Body.Close()

	// Handle possible gzip encoding
	var reader io.Reader = resp.Body
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		var gzipReader *gzip.Reader
		if gzipReader, err = gzip.NewReader(reader); err != nil {
			return a.logError("Error creating gzip reader:", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Read the response body
	bodyContent, totalSize := a.readResponseBody(reader)

	contentType := resp.Header.Get("Content-Type")
	size := formatSize(totalSize)
	output := handleResponseContent(contentType, bodyContent)
	roundedTime := math.Round(time.Since(start).Seconds()*1000*100) / 100

	return Response{
		Response:  output,
		Status:    resp.Status,
		Size:      size,
		TotalTime: roundedTime,
		Headers:   resp.Header,
	}
}

func (a *App) logError(message string, err error) Response {
	fmt.Println(message, err)
	return Response{}
}

func (a *App) readResponseBody(reader io.Reader) ([]byte, int) {
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
			a.logError("Error reading the response body:", err)
			break
		}
	}

	return []byte(bodyContent.String()), totalSize
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

func (a *App) MinimizeWindow() {
	runtime.WindowMinimise(a.ctx)
}
func (a *App) MaximizeWindow() {
	runtime.WindowToggleMaximise(a.ctx)
}
func (a *App) CloseWindow() {
	runtime.Quit(a.ctx)
}
