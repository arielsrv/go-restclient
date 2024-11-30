package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

type SiteResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CountryID   string `json:"country_id,omitempty"`
}

type CountryResponse struct {
	Id                 string `json:"id"`
	Name               string `json:"name"`
	Locale             string `json:"locale"`
	CurrencyId         string `json:"currency_id"`
	DecimalSeparator   string `json:"decimal_separator"`
	ThousandsSeparator string `json:"thousands_separator"`
	TimeZone           string `json:"time_zone"`
	TimeZoneName       string `json:"time_zone_name"`
}

func main() {
	client := &rest.Client{
		Name:           "sites-client",
		BaseURL:        "https://api.prod.dp.iskaypet.com",
		Timeout:        time.Millisecond * time.Duration(2000),
		ConnectTimeout: time.Millisecond * time.Duration(5000),
		ContentType:    rest.JSON,
	}

	ctx := context.Background()
	response := client.GetWithContext(ctx, "/sites")
	if response.Err != nil {
		log.Fatal(response.Err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.String())
	}

	log.Infof("Cache-Control: %v", response.Header.Get("Cache-Control"))

	// Typed fill up
	sitesResponse, err := rest.Deserialize[[]SiteResponse](response)
	if err != nil {
		log.Fatal(err)
	}

	for i := range sitesResponse {
		log.Infof("Site: %v", sitesResponse[i])

		for k := range sitesResponse {
			response = client.GetWithContext(ctx, fmt.Sprintf("/sites/%s", sitesResponse[k].ID))
			if response.Err != nil {
				log.Fatal(response.Err)
			}

			if response.StatusCode != http.StatusOK {
				log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.String())
			}

			siteResponse, sErr := rest.Deserialize[SiteResponse](response)
			if sErr != nil {
				log.Fatal(sErr)
			}

			log.Infof("Site Details: %v", siteResponse)

			response = client.GetWithContext(ctx, fmt.Sprintf("/countries/%s", siteResponse.CountryID))
			if response.Err != nil {
				log.Fatal(response.Err)
			}

			if response.StatusCode != http.StatusOK {
				log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.String())
			}

			countryResponse, cErr := rest.Deserialize[CountryResponse](response)
			if cErr != nil {
				log.Fatal(cErr)
			}

			log.Infof("Country Details: %v", countryResponse)
		}
	}

	// Cache-Control (hit)
	response = client.GetWithContext(ctx, "/sites")
	switch {
	case response.Err != nil:
		log.Fatal(response.Err)
	case response.StatusCode != http.StatusOK:
		log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.String())
	}

	log.Infof("Cache-Control: %v", response.Header.Get("Cache-Control"))
}
