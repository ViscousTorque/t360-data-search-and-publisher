# t360-data-search-and-publisher

## Developer Test: API Data Search and Pub/Sub Publisher
* QUESTION: what is expected when all endpoints fail to respond within timeout, there is no published positive result
* TODO: build test_search mocks for testing, stop using sandbox
* TODO: add a global rate limiter as a crude way to prevent endpoints from being overwhelmed.
* TODO: Need to provide mechanism in docker to work on other platforms / arch
* TODO: metrics and graceful shutdown on SIGTERMs?
* TODO: write unittests?
* TODO: repackage, refactor, refactor, refactor
* TODO: external queue for resilience, instead of go channel ... asyncd & redis?

### Objective
Create a Go application that:
* Post data to multiple REST API endpoints creating a unique reference to the search, if any endpoint takes longer than 2 seconds ignore the result.
* Publishes the processed data to a Google Cloud Pub/Sub topic.

### Requirements
* Language: Go (latest stable version).
* Data Sources: The following API endpoints, which post below search body JSON packet:
    * https://sandbox-update.transfer360.dev/test_search/acmelease
    * https://sandbox-update.transfer360.dev/test_search/leasecompany
    * https://sandbox-update.transfer360.dev/test_search/fleetcompany
    * https://sandbox-update.transfer360.dev/test_search/hirecompany

Output: Simulate publishing positive hire vehicles results to a Pub/Sub topic named: positive_searches using the JSON format below pub/sub packet

### Deliverables

* A GitHub repository with the full source code.
* A README.md file explaining how to run the application locally.

## Data Information

Search Body

```
{
 "vrm":"TEST123",
 "contravention_date":"2025-03-17T11:15:20Z"
}
```

Pub/Sub Packet
```
{
  "reference": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "vrm": "TE10EST",
  "contravention_date": "2022-05-23T11:12:33Z",
  "is_hirer_vehicle": true,
  "lease_company": {
    "companyname": "Acme Fleet Hire Limited",
    "address_line1": "Address Line 1",
    "address_line2": "Address Line 2",
    "addres_line3": "Address Line 3",
    "addres_line4": "Address Line 4",
    "postcode": "post code"
  }
}
```

### Notes
A list of test vehicles can be found on this link: https://sandbox-update.transfer360.dev/test_vehicles

## Environment so far

### VS Code
IDE vs code with relevant plugins:
* Python
* Python Debugger
* Pylance

### Ubuntu 24.04 desktop

## Further reading
* Go packages : 
    * https://pkg.go.dev/github.com/PuerkitoBio/goquery decoding / scraping the html test data webpage
        * https://www.slingacademy.com/article/fetching-and-parsing-html-pages-with-goquery/
    * Google Pub Sub
        * https://cloud.google.com/pubsub/docs/overview
        * https://pkg.go.dev/cloud.google.com/go/pubsub

## Testing on dev desktop
docker compose up --build --abort-on-container-exit

## Run with go for local testing
Run the fake sub and then then the go app

```
./start-emulator.sh

Starting Pub/Sub Emulator container...
b471af229fd26d1d00c2adf1f83c7886cb9f42c358c7143ed9bdd7599fb08def
```

Then run the go command:
```
PUBSUB_EMULATOR_HOST=localhost:8085 go run main.go
```

You can stop the emulator with:
```
docker stop pubsub-emulator
```

## Vs Code Run
Use the Run and Debug to "Launch Go App" or press F5

## Run instructions with docker
Update the env variables incase you need to use something other than default

docker run --rm \
  -e PUBSUB_EMULATOR_HOST=localhost:8085 \
  -e PUBSUB_PROJECT_ID=my-gcp-project \
  -e GCP_PROJECT_ID=my-gcp-project \
  -e PUBSUB_TOPIC=my-pubsub-topic \
  -e UBSUB_EMULATOR_HOST=localhost:8085 \
  t360-data-search-and-publisher-app

All config env variables are set with defaults, see app log:
* VEHICLE_LIST_URL = "https://sandbox-update.transfer360.dev/test_vehicles"
* SEARCH_APIS = "https://sandbox-update.transfer360.dev/test_search/acmelease","https://sandbox-update.transfer360.dev/test_search/fleetcompany",
			"https://sandbox-update.transfer360.dev/test_search/hirecompany","https://sandbox-update.transfer360.dev/test_search/leasecompany"
* REQUEST_TIMEOUT = 2s
* GCP_PROJECT_ID = "my-gcp-project"
* PUBSUB_TOPIC = "my-pubsub-topic"
* PUBSUB_EMULATOR_HOST = ""
