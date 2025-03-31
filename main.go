package main

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"main/config"
	"main/gcpubsub"
	"main/search"
	"main/vehicle"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
)

func main() {
	log.Println("Starting application")

	if config.PubSubEmulatorHost != "" {
		log.Println("Using Pub/Sub Emulator at:", config.PubSubEmulatorHost)
	}

	vehicles, err := vehicle.FetchVehicleList()
	if err != nil {
		log.Fatalf("Error fetching vehicle list: %v", err)
	}

	// only need one Client for all pubsubs
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, config.GCPProjectID)
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
	}
	defer client.Close()

	workerCount, err := strconv.Atoi(config.WorkerCount)
	if err != nil {
		log.Fatalf("Failed to get worker count: %v", err)
	}
	log.Printf("Using %v workers", workerCount)

	jobs := make(chan vehicle.Vehicle)
	var wg sync.WaitGroup

	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go worker(i, client, jobs, &wg)
	}

	for _, v := range vehicles {
		jobs <- v
	}
	close(jobs)

	wg.Wait()
	log.Println("Finished processing vehicles data.")
}

func worker(id int, client *pubsub.Client, jobs <-chan vehicle.Vehicle, wg *sync.WaitGroup) {
	defer wg.Done()

	for v := range jobs {
		searchRef := uuid.New().String()
		log.Printf("[Worker %d] Starting search for VRM: %s (Ref: %s)", id, v.VRM, searchRef)

		results := search.PerformSearch(v.VRM)

		if len(results) == 0 {
			log.Printf("[Worker %d] No results for %s, skipping Pub/Sub.", id, v.VRM)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err := gcpubsub.PublishToPubSub(ctx, client, searchRef, results)
		cancel()

		if err != nil {
			log.Printf("[Worker %d] Publish error for %s: %v", id, v.VRM, err)
		} else {
			log.Printf("[Worker %d] Successfully published results for %s", id, v.VRM)
		}
	}
}
