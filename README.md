# t360-data-search-and-publisher

## Developer Test: API Data Search and Pub/Sub Publisher

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

### Notes
A list of test vehicles can be found on this link: https://sandbox-update.transfer360.dev/test_vehicles

## Environment so far

### VS Code
IDE vs code with relevant plugins:
* Python
* Python Debugger
* Pylance

### Ubuntu 24.04 desktop
* TODO: Need to provide mechanism in docker to work on other platforms / arch

## Further reading
* Go packages : 
    * https://pkg.go.dev/github.com/PuerkitoBio/goquery decoding / scraping the html test data webpage
        * https://www.slingacademy.com/article/fetching-and-parsing-html-pages-with-goquery/
    * Google Pub Sub
        * https://cloud.google.com/pubsub/docs/overview
        * https://pkg.go.dev/cloud.google.com/go/pubsub


