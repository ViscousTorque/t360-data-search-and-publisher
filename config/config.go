package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Printf("Env variable %s not set or empty, using default: %s", key, defaultValue)
		return defaultValue
	}
	return value
}

func GetEnvInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Printf("Env variable %s not set or empty, using default: %d", key, defaultValue)
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Invalid integer for %s: %s, using default: %d", key, value, defaultValue)
		return defaultValue
	}
	return intValue
}

func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Printf("Env variable %s not set or empty, using default: %s", key, defaultValue)
		return defaultValue
	}
	durationValue, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("Invalid duration for %s: %s, using default: %s", key, value, defaultValue)
		return defaultValue
	}
	return durationValue
}

func GetEnvList(key string, defaultValue []string) []string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Printf("Env variable %s not set or empty, using default: %v", key, defaultValue)
		return defaultValue
	}
	return strings.Split(value, ",")
}

func LookupEnvStrict(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Fatalf("Missing required environment variable: %s", key)
	}
	return value
}

var (
	VehicleListURL = GetEnv("VEHICLE_LIST_URL", "https://sandbox-update.transfer360.dev/test_vehicles")
	SearchAPIs     = GetEnvList("SEARCH_APIS",
		[]string{"https://sandbox-update.transfer360.dev/test_search/acmelease",
			"https://sandbox-update.transfer360.dev/test_search/fleetcompany",
			"https://sandbox-update.transfer360.dev/test_search/hirecompany",
			"https://sandbox-update.transfer360.dev/test_search/leasecompany"})
	RequestTimeout     = GetEnvDuration("REQUEST_TIMEOUT", 5*time.Second)
	GCPProjectID       = GetEnv("GCP_PROJECT_ID", "your-gcp-project-id")
	PubSubTopic        = GetEnv("PUBSUB_TOPIC", "your-pubsub-topic")
	PubSubEmulatorHost = GetEnv("PUBSUB_EMULATOR_HOST", "")
)
