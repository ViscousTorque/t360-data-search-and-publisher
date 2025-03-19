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
	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()

	payload := SearchRequest{
		VRM:               vrm,
		ContraventionDate: time.Now().Format(time.RFC3339),
	}

	var wg sync.WaitGroup
	resultsCh := make(chan *APIResponse, len(config.SearchAPIs))

	for _, endpointUrl := range config.SearchAPIs {
		if !isValidURL(endpointUrl) {
			log.Printf("Invalid URL: %s", endpointUrl)
			continue
		}

		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			response := sendRequest(ctx, endpointUrl, payload)
			select {
			case resultsCh <- response:
			case <-ctx.Done():
				log.Printf("Timeout occurred, ignoring response from %s", url)
			}
		}(endpointUrl)
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

func isValidURL(rawURL string) bool {
	parsedURL, err := url.ParseRequestURI(rawURL)
	return err == nil && parsedURL.Scheme != "" && parsedURL.Host != ""
}

// PerformSearch sends requests to multiple APIs concurrently
func PerformSearchOrig(vrm string) []APIResponse {
	log.Println("Perform Search")
	ctx, cancel := context.WithTimeout(context.Background(), config.RequestTimeout)
	defer cancel()

	payload := SearchRequest{
		VRM:               vrm,
		ContraventionDate: time.Now().Format(time.RFC3339),
	}

	ch := make(chan *APIResponse, len(config.SearchAPIs))
	var wg sync.WaitGroup

	for _, endpoint := range config.SearchAPIs {
		if !isValidURL(endpoint) {
			log.Printf("Invalid URL: %s", endpoint)
			continue
		}
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			data, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Error marshalling payload: %v", err)
				ch <- &APIResponse{Endpoint: url, Error: "JSON marshalling failed"}
				return
			}

			req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
			if err != nil {
				log.Printf("Error creating HTTP request: %v", err)
				ch <- &APIResponse{Endpoint: url, Error: "Failed to create request"}
				return
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			log.Println(resp)

			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) {
					log.Printf("Network error making request to %s: %v", url, err)
					ch <- &APIResponse{Endpoint: url, Error: "Network error: " + err.Error()}
					return
				}
				log.Printf("Error making request to %s: %v", url, err)
				ch <- &APIResponse{Endpoint: url, Error: err.Error()}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Non-200 response from %s: %d", url, resp.StatusCode)
				ch <- &APIResponse{Endpoint: url, Error: "HTTP error: " + http.StatusText(resp.StatusCode)}
				return
			}

			select {
			case ch <- &APIResponse{Endpoint: url, Data: "Success"}:
			case <-ctx.Done():
				log.Printf("Timeout occurred, ignoring response from %s", url)
			}
		}(endpoint)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var results []APIResponse
	for res := range ch {
		if res != nil {
			results = append(results, *res)
		}
	}

	return results
}
