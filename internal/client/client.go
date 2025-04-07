package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/NETWAYS/go-check"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Client struct {
	URL          string
	Client       api.Client
	API          v1.API
	RoundTripper http.RoundTripper
}

func NewClient(url string, rt http.RoundTripper) *Client {
	return &Client{
		URL:          url,
		RoundTripper: rt,
	}
}

// nolint: gosec
func (c *Client) Connect() error {
	cfg, err := api.NewClient(api.Config{
		Address:      c.URL,
		RoundTripper: c.RoundTripper,
	})

	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	c.Client = cfg
	c.API = v1.NewAPI(c.Client)

	return nil
}

func (c *Client) GetStatus(ctx context.Context, endpoint string) (returncode int, statuscode int, body string, err error) {
	// Parses the response from the Prometheus /healthy and /ready endpoint
	// Return: Exit Status Code, HTTP Status Code, HTTP Body, Error
	// Building the final URL with the endpoint parameter
	u, _ := url.JoinPath(c.URL, "/-/", endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)

	if err != nil {
		e := fmt.Sprintf("could not create request: %s", err)
		return check.Unknown, 0, e, err
	}

	// Making the request with the preconfigured Client
	// So that we can reuse the preconfigured Roundtripper
	resp, b, err := c.Client.Do(ctx, req)

	if err != nil {
		e := fmt.Sprintf("could not get status: %s", err)
		return check.Unknown, 0, e, err
	}

	defer resp.Body.Close()

	// Getting the response body
	respBody := strings.TrimSpace(string(b))

	// What we expect from the Prometheus Server
	statusOk := "is Healthy."
	if endpoint == "ready" {
		statusOk = "is Ready."
	}

	if resp.StatusCode == http.StatusOK && strings.Contains(respBody, statusOk) {
		return check.OK, resp.StatusCode, respBody, err
	}

	if resp.StatusCode != http.StatusOK {
		return check.Critical, resp.StatusCode, respBody, err
	}

	return check.Unknown, resp.StatusCode, respBody, err
}

type headersRoundTripper struct {
	headers map[string]string
	rt      http.RoundTripper
}

// NewHeadersRoundTripper adds the given headers to a request
func NewHeadersRoundTripper(headers map[string]string, rt http.RoundTripper) http.RoundTripper {
	return &headersRoundTripper{headers, rt}
}

func (rt *headersRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// RoundTrip should not modify the request, except for
	// consuming and closing the Request's Body.
	req = cloneRequest(req)

	for key, value := range rt.headers {
		req.Header.Add(key, value)
	}

	return rt.rt.RoundTrip(req)
}

// cloneRequest returns a clone of the provided *http.Request
func cloneRequest(r *http.Request) *http.Request {
	// Shallow copy of the struct.
	r2 := new(http.Request)
	*r2 = *r
	// Deep copy of the Header.
	r2.Header = make(http.Header)
	for k, s := range r.Header {
		r2.Header[k] = s
	}

	return r2
}
