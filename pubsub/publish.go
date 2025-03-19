package pubsub

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"main/config"
	"main/search"

	"cloud.google.com/go/pubsub"
)

type VehicleData struct {
	VRM               string `json:"vrm"`
	ContraventionDate string `json:"contravention_date"`
	IsHirerVehicle    bool   `json:"is_hirer_vehicle"`
	LeaseCompany      any    `json:"lease_company"`
}

func filterHirerVehicles(results []search.APIResponse) []search.APIResponse {
	var filteredResults []search.APIResponse

	for _, result := range results {
		if result.Error != "" {
			continue
		}

		var vehicleData VehicleData
		if err := json.Unmarshal([]byte(result.Data), &vehicleData); err != nil {
			log.Printf("Skipping invalid JSON from %s: %v", result.Endpoint, err)
			continue
		}

		if vehicleData.IsHirerVehicle {
			filteredResults = append(filteredResults, result)
		}
	}

	return filteredResults
}

func waitForPubSubTopic(ctx context.Context, client *pubsub.Client, topicName string, maxRetries int, delay time.Duration) {
	topic := client.Topic(topicName)

	for i := 0; i < maxRetries; i++ {
		exists, err := topic.Exists(ctx)
		if err == nil && exists {
			log.Printf("Pub/Sub topic %s is ready.", topicName)
			return
		}
		log.Printf("Waiting for Pub/Sub topic %s to be available... (%d/%d)", topicName, i+1, maxRetries)
		time.Sleep(delay)
	}

	log.Fatalf("Pub/Sub topic %s did not become available in time.", topicName)
}

// PublishToPubSub sends processed data to a Pub/Sub topic
func PublishToPubSub(searchRef string, results []search.APIResponse) {
	ctx := context.Background()

	if config.PubSubEmulatorHost != "" {
		log.Println("Using Pub/Sub Emulator at:", config.PubSubEmulatorHost)
		os.Setenv("PUBSUB_EMULATOR_HOST", config.PubSubEmulatorHost)
	}

	client, err := pubsub.NewClient(ctx, config.GCPProjectID)
	if err != nil {
		log.Fatalf("Error creating Pub/Sub client: %v", err)
	}
	defer client.Close()

	waitForPubSubTopic(ctx, client, config.PubSubTopic, 10, 2*time.Second)

	filteredResults := filterHirerVehicles(results)

	if len(filteredResults) == 0 {
		log.Println("No valid results to publish.")
		return
	}

	message, err := json.Marshal(map[string]any{
		"reference": searchRef,
		"results":   filteredResults,
	})
	if err != nil {
		log.Fatalf("Error marshalling Pub/Sub message: %v", err)
	}

	topic := client.Topic(config.PubSubTopic)
	result := topic.Publish(ctx, &pubsub.Message{Data: message})
	_, err = result.Get(ctx)
	if err != nil {
		log.Fatalf("Error publishing to Pub/Sub: %v", err)
	}

	log.Printf("Published search reference %s to Pub/Sub.", searchRef)
}
