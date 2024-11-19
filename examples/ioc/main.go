package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

type SiteResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ISitesClient interface {
	GetSites(ctx context.Context) ([]SiteResponse, error)
}

type SitesClient struct {
	client rest.HTTPClient
}

func NewSitesClient(
	client rest.HTTPClient,
) *SitesClient {
	return &SitesClient{
		client: client,
	}
}

func (r SitesClient) GetSites(ctx context.Context) ([]SiteResponse, error) {
	var sitesResponse []SiteResponse

	response := r.client.GetWithContext(ctx, "/sites")
	if response.Err != nil {
		return nil, response.Err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", response.StatusCode, response.String())
	}

	err := response.FillUp(&sitesResponse)
	if err != nil {
		return nil, err
	}

	return sitesResponse, nil
}

func main() {
	ctx := context.Background()

	var client rest.HTTPClient = &rest.Client{
		Name:        "sites-client",
		BaseURL:     "https://api.prod.dp.iskaypet.com",
		ContentType: rest.JSON,
		Timeout:     time.Duration(5000) * time.Millisecond,
	}

	var sitesClient ISitesClient = NewSitesClient(client)

	sites, err := sitesClient.GetSites(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for i := range sites {
		fmt.Printf("Site: %+v\n", sites[i])
	}
}
