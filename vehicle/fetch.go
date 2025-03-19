package vehicle

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"main/config"

	"github.com/PuerkitoBio/goquery"
)

type Vehicle struct {
	VRM         string
	HireCompany string
}

// FetchVehicleList: scrapes the sandbox testlist webpage and extracts vehicle data
func FetchVehicleList() ([]Vehicle, error) {
	resp, err := http.Get(config.VehicleListURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET Vehicle List Error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var vehicles []Vehicle

	doc.Find("table tbody tr").Each(func(index int, row *goquery.Selection) {
		columns := row.Find("td")
		if columns.Length() >= 2 {
			vrm := strings.TrimSpace(columns.Eq(0).Text())
			hireCompany := strings.TrimSpace(columns.Eq(1).Text())

			vehicles = append(vehicles, Vehicle{VRM: vrm, HireCompany: hireCompany})
		}
	})

	if len(vehicles) == 0 {
		log.Println("No vehicles found in the webpage response")
	}

	log.Printf("Fetched %d vehicles.", len(vehicles))
	return vehicles, nil
}
