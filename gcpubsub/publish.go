package gcpubsub

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"main/config"
	"main/models"
	"main/search"

	"cloud.google.com/go/pubsub"
)

type PublishedHirerVehicle struct {
	Reference string `json:"reference"`
	models.VehicleData
}

func filterHirerVehicles(results []search.APIResponse) (hirerVehicles []search.APIResponse, skippedErrors []search.APIResponse) {
	for _, result := range results {
		if result.Error != "" {
			log.Printf("Skipping errored response from %s: %s", result.Endpoint, result.Error)
			skippedErrors = append(skippedErrors, result)
			continue
		}

		var vehicleData models.VehicleData
		if err := json.Unmarshal([]byte(result.Data), &vehicleData); err != nil {
			log.Printf("Skipping invalid JSON from %s: %v", result.Endpoint, err)
			skippedErrors = append(skippedErrors, result)
			continue
		}

		if vehicleData.IsHirerVehicle {
			hirerVehicles = append(hirerVehicles, result)
		}
	}
	log.Printf("Hirer vehicles %v and skipped Errors %v", hirerVehicles, skippedErrors)

	return hirerVehicles, skippedErrors
}

func waitForPubSubTopic(ctx context.Context, client *pubsub.Client, topicName string, maxRetries int, delay time.Duration) {
	topic := client.Topic(topicName)

	for i := range maxRetries {
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
func PublishToPubSub(ctx context.Context, client *pubsub.Client, searchRef string, results []search.APIResponse) error {
	waitForPubSubTopic(ctx, client, config.PubSubTopic, 10, 2*time.Second)

	hirers, skipped := filterHirerVehicles(results)
	log.Printf("Skipped %d responses", len(skipped))

	if len(hirers) == 0 {
		log.Println("No valid results to publish.")
		return nil
	}

	var vehicle models.VehicleData
	if err := json.Unmarshal([]byte(hirers[0].Data), &vehicle); err != nil {
		log.Fatalf("Invalid VehicleData in APIResponse from %s: %v", hirers[0].Endpoint, err)
	}

	msg := PublishedHirerVehicle{
		Reference:   searchRef,
		VehicleData: vehicle,
	}

	message, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("error marshalling Pub/Sub message: %w", err)
	}

	topic := client.Topic(config.PubSubTopic)
	result := topic.Publish(ctx, &pubsub.Message{Data: message})

	if _, err := result.Get(ctx); err != nil {
		log.Fatalf("Error publishing to Pub/Sub: %v", err)
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, message, "", "  "); err != nil {
		log.Printf("Failed to pretty-print message: %v", err)
	}

	log.Printf("Published search reference %s to Pub/Sub.", searchRef)
	log.Printf("Publishing message:\n%s", pretty.String())

	return nil
}
