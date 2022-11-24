package client

import (
	"context"
	"fmt"
	"github.com/NETWAYS/go-check"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	Url          string
	Client       api.Client
	Api          v1.API
	RoundTripper http.RoundTripper
}

func NewClient(url string, rt http.RoundTripper) *Client {
	return &Client{
		Url:          url,
		RoundTripper: rt,
	}
}

// nolint: gosec
func (c *Client) Connect() error {
	cfg, err := api.NewClient(api.Config{
		Address:      c.Url,
		RoundTripper: c.RoundTripper,
	})

	if err != nil {
		return fmt.Errorf("Error creating client: %w\n", err)
	}

	c.Client = cfg
	c.Api = v1.NewAPI(c.Client)

	return nil
}

func (c *Client) GetStatus(ctx context.Context, endpoint string) (int, string, error) {
	// Parses the response from the Prometheus /healthy and /ready endpoint

	// Building the final URL with the endpoint parameter
	u, _ := url.JoinPath(c.Url, "/-/", endpoint)
	req, err := http.NewRequest(http.MethodGet, u, nil)

	if err != nil {
		e := fmt.Sprintf("Could not create request: %s", err)
		return check.Unknown, e, err
	}

	// Making the request with the preconfigured Client
	// So that we can reuse the preconfigured Roundtripper
	resp, body, err := c.Client.Do(ctx, req)

	if err != nil {
		e := fmt.Sprintf("Could get status: %s", err)
		return check.Unknown, e, err
	}

	// Getting the response body
	respBody := strings.TrimSpace(string(body))

	// What we expect from the Prometheus Server
	statusOk := "Prometheus Server is Healthy."
	if endpoint == "ready" {
		statusOk = "Prometheus Server is Ready."
	}

	if resp.StatusCode == 200 && respBody == statusOk {
		return check.OK, respBody, err
	}

	if resp.StatusCode != 200 {
		return check.Critical, respBody, err
	}

	return check.Unknown, respBody, err
}
