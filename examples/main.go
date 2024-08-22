package main

import (
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
	baseURL := "https://staging-eu01-kiwoko.demandware.net/s/-/dw/data/v22_6"

	rb := rest.RequestBuilder{
		Timeout:        time.Millisecond * 500,
		ConnectTimeout: time.Millisecond * 2000,
		BaseURL:        baseURL,
		OAuth: &clientcredentials.Config{
			ClientID:     "a11d0149-687e-452e-9c94-783d489d4f72",
			ClientSecret: "Kiwoko@1234",
			TokenURL:     "https://account.demandware.com/dw/oauth2/access_token",
			AuthStyle:    oauth2.AuthStyleInHeader,
		},
	}

	var sitesResponse struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	for {
		response := rb.Get("/sites")
		if response.Err != nil {
			log.Info(response.Err)
			continue
		}

		if response.StatusCode != http.StatusOK {
			log.Info(response.String())
			log.Infof("invalid status_code: %d", response.StatusCode)
			continue
		}

		err := response.FillUp(&sitesResponse)
		if err != nil {
			log.Info(err)
			continue
		}

		log.Info("Sites: ")
		for i := 0; i < len(sitesResponse.Data); i++ {
			log.Infof("\t%s", sitesResponse.Data[i].ID)
		}

		time.Sleep(1000 * time.Millisecond)
	}
}
