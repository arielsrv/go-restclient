package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	var (
		httpClient  rest.HTTPClient
		sitesClient ISitesClient
	)

	httpClient = &rest.Client{
		Name:        "sitesResponse-httpClient",
		BaseURL:     "https://api.prod.dp.iskaypet.com",
		ContentType: rest.JSON,
		Timeout:     time.Duration(5000) * time.Millisecond,
	}

	sitesClient = NewSitesClient(httpClient)

	sitesResponse, err := sitesClient.GetSites(context.Background())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for i := range sitesResponse {
		fmt.Printf("Site: %+v\n", sitesResponse[i])
	}
}

type SiteResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ISitesClient interface {
	GetSites(ctx context.Context) ([]SiteResponse, error)
}

type SitesClient struct {
	httpClient rest.HTTPClient
}

func NewSitesClient(
	httpClient rest.HTTPClient,
) *SitesClient {
	return &SitesClient{
		httpClient: httpClient,
	}
}

func (r SitesClient) GetSites(ctx context.Context) ([]SiteResponse, error) {
	apiURL := "/sites"

	headers := make(http.Header)
	headers.Set("x-api-key", "your-api-key")

	response := r.httpClient.GetWithContext(ctx, apiURL, headers)
	if response.Err != nil {
		return nil, response.Err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", response.StatusCode, response.String())
	}

	var sitesResponse []SiteResponse
	err := response.FillUp(&sitesResponse)
	if err != nil {
		return nil, err
	}

	return sitesResponse, nil
}
