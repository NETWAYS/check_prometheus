package client

import (
	"crypto/tls"
	"fmt"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"net/http"
)

type Client struct {
	Url      string
	Client   api.Client
	Api      v1.API
	Insecure bool
}

func NewClient(url string) *Client {
	return &Client{
		Url:      url,
		Insecure: false,
	}
}

// nolint: gosec
func (c *Client) Connect() error {
	var rt = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.Insecure},
	}

	cfg, err := api.NewClient(api.Config{
		Address:      c.Url,
		RoundTripper: rt,
	})

	if err != nil {
		return fmt.Errorf("Error creating client: %w\n", err)
	}

	c.Client = cfg
	c.Api = v1.NewAPI(c.Client)

	return nil
}
