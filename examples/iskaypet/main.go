package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"gitlab.com/arielsrv/go-restclient/rest"
)

type SiteResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CountryID   string `json:"country_id,omitempty"`
}

type CountryResponse struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Locale             string `json:"locale"`
	CurrencyID         string `json:"currency_id"`
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
		fmt.Println(response.Err)
		os.Exit(1)
	}

	if response.StatusCode != http.StatusOK {
		fmt.Printf("Status: %d, Body: %s\n", response.StatusCode, response.String())
		os.Exit(1)
	}

	fmt.Printf("Cache-Control: %v\n", response.Header.Get("Cache-Control"))

	// Typed fill up
	sitesResponse, err := rest.Deserialize[[]SiteResponse](response)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for i := range sitesResponse {
		fmt.Printf("Site: %v\n", sitesResponse[i])

		for k := range sitesResponse {
			response = client.GetWithContext(ctx, fmt.Sprintf("/sites/%s", sitesResponse[k].ID))
			if response.Err != nil {
				fmt.Println(response.Err)
				os.Exit(1)
			}

			if response.StatusCode != http.StatusOK {
				fmt.Printf("Status: %d, Body: %s\n", response.StatusCode, response.String())
				os.Exit(1)
			}

			siteResponse, sErr := rest.Deserialize[SiteResponse](response)
			if sErr != nil {
				fmt.Println(sErr)
				os.Exit(1)
			}

			fmt.Printf("Site Details: %v\n", siteResponse)

			response = client.GetWithContext(ctx, fmt.Sprintf("/countries/%s", siteResponse.CountryID))
			if response.Err != nil {
				fmt.Println(response.Err)
				os.Exit(1)
			}

			if response.StatusCode != http.StatusOK {
				fmt.Printf("Status: %d, Body: %s\n", response.StatusCode, response.String())
				os.Exit(1)
			}

			countryResponse, cErr := rest.Deserialize[CountryResponse](response)
			if cErr != nil {
				fmt.Println(cErr)
				os.Exit(1)
			}

			fmt.Printf("Country Details: %v\n", countryResponse)
		}
	}

	// Cache-Control (hit)
	response = client.GetWithContext(ctx, "/sites")
	switch {
	case response.Err != nil:
		fmt.Println(response.Err)
		os.Exit(1)
	case response.StatusCode != http.StatusOK:
		fmt.Printf("Status: %d, Body: %s\n", response.StatusCode, response.String())
		os.Exit(1)
	}

	fmt.Printf("Cache-Control: %v\n", response.Header.Get("Cache-Control"))
}
