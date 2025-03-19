package main

import (
	"encoding/json"
	"log"

	"main/pubsub"
	"main/search"
	"main/vehicle"

	"github.com/google/uuid"
)

func main() {
	log.Println("Starting application")

	vehicles, err := vehicle.FetchVehicleList()
	if err != nil {
		log.Fatalf("Error fetching vehicle list: %v", err)
	}

	for _, vehicle := range vehicles {
		searchRef := uuid.New().String()
		log.Printf("Processing search for VRM: %s (Ref: %s)", vehicle.VRM, searchRef)

		results := search.PerformSearch(vehicle.VRM)
		jsonData, err := json.Marshal(results)
		if err != nil {
			log.Printf("Error marshalling results: %v", err)
		} else {
			log.Printf("Results collected: %s", jsonData)
		}

		if len(results) > 0 {
			pubsub.PublishToPubSub(searchRef, results)
		} else {
			log.Println("No valid results found, skipping Pub/Sub publishing.")
		}

	}

	log.Println("Finished processing vehicles data.")
}
