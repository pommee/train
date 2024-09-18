package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}
	}

	contentType := resp.Header.Get("Content-Type")
	var prettyBody string

	if contentType == "application/json" {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, body, "", "  ")
		if err != nil {
			prettyBody = string(body)
		} else {
			prettyBody = prettyJSON.String()
		}
	} else {
		prettyBody = string(body)
	}

	roundedTime := math.Round(time.Since(start).Seconds()*1000*100) / 100

	return Response{
		Response:  prettyBody,
		Status:    resp.Status,
		Size:      fmt.Sprintf("%d bytes", len(body)),
		TotalTime: roundedTime,
		Headers:   resp.Header,
	}
}

func (a *App) logError(message string, err error) Response {
	fmt.Println(message, err)
	return Response{}
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
