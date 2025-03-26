package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"main/config"
	"main/models"
)

type SearchRequest struct {
	VRM               string `json:"vrm"`
	ContraventionDate string `json:"contravention_date"`
}

type APIResponse struct {
	Endpoint string `json:"endpoint"`
	Data     string `json:"data"`
	Error    string `json:"error,omitempty"`
}

// PerformSearch sends requests to multiple APIs concurrently
func PerformSearch(vrm string) []APIResponse {
	ctx, timeoutCancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer timeoutCancel()

	payload := SearchRequest{
		VRM:               vrm,
		ContraventionDate: time.Now().Format(time.RFC3339),
	}

	var wg sync.WaitGroup
	resultsCh := make(chan *APIResponse, len(config.SearchAPIs))
	stopCh := make(chan struct{})

	for _, endpoint := range config.SearchAPIs {
		url := endpoint
		if !isValidURL(endpoint) {
			log.Printf("Invalid URL: %s", endpoint)
			continue
		}

		select {
		case <-stopCh:
			log.Printf("Skipping request to %s — true hire already found", url)
			continue
		default:
		}

		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			response := sendRequest(ctx, url, payload)

			if response == nil {
				return
			}

			if isTrueHire(response) {
				log.Printf("True hire from %s — stopping future launches", url)
				select {
				case <-stopCh:
				default:
					close(stopCh)
				}

				// Only push to resultsCh if it's a true hire
				select {
				case <-ctx.Done():
					log.Printf("Context cancelled after request to %s", url)
					return
				case resultsCh <- response:
				}
			} else {
				log.Printf("Non-hirer result from %s — not sending to results channel", url)
			}
		}(url)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var results []APIResponse
	for res := range resultsCh {
		if res != nil {
			results = append(results, *res)
		}
	}

	return results
}

func isValidURL(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

func isTrueHire(resp *APIResponse) bool {
	if resp == nil || resp.Error != "" {
		return false
	}

	var data models.VehicleData
	if err := json.Unmarshal([]byte(resp.Data), &data); err != nil {
		log.Printf("Invalid JSON from %s: %v", resp.Endpoint, err)
		return false
	}

	return data.IsHirerVehicle
}

// sendRequest handles the HTTP request logic
func sendRequest(ctx context.Context, url string, payload SearchRequest) *APIResponse {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling payload: %v", err)
		return &APIResponse{Endpoint: url, Error: "JSON marshalling failed"}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return &APIResponse{Endpoint: url, Error: "Failed to create request"}
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			log.Printf("Request to %s was cancelled or timed out", url)
			return &APIResponse{
				Endpoint: url,
				Error:    "Request cancelled or timed out",
			}
		}

		var netErr net.Error
		if errors.As(err, &netErr) {
			log.Printf("Network error making request to %s: %v", url, err)
			return &APIResponse{Endpoint: url, Error: "Network error: " + err.Error()}
		}
		log.Printf("Error making request to %s: %v", url, err)
		return &APIResponse{Endpoint: url, Error: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-200 response from %s: %d", url, resp.StatusCode)
		return &APIResponse{Endpoint: url, Error: "HTTP error: " + http.StatusText(resp.StatusCode)}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIResponse{Endpoint: url, Error: err.Error()}
	}

	return &APIResponse{Endpoint: url, Data: string(body)}
}
