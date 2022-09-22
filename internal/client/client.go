package client

import (
	"fmt"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Client struct {
	Url    string
	Client api.Client
	Api    v1.API
}

func NewClient(url string) *Client {
	return &Client{
		Url: url,
	}
}

// nolint: gosec
func (c *Client) Connect() error {
	cfg, err := api.NewClient(api.Config{
		Address: c.Url,
	})

	if err != nil {
		fmt.Errorf("Error creating client: %v\n", err)
		return err
	}

	// TODO Test connection?
	c.Client = cfg
	c.Api = v1.NewAPI(c.Client)

	return nil
}
