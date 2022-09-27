package client

import (
	"crypto/tls"
	"fmt"
	"github.com/NETWAYS/go-check"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"io/ioutil"
	"net/http"
	"strings"
)

type Client struct {
	Url      string
	Client   api.Client
	Api      v1.API
	Insecure bool
}

const (
	HealthGood = "Prometheus Server is Healthy."
	ReadyGood  = "Prometheus Server is Ready."
)

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

func (c *Client) GetStatus(response *http.Response) (rc int, output string, err error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return check.Unknown, "", err
	}

	output = strings.TrimSpace(string(body))

	if response.StatusCode == 200 && (output == HealthGood || output == ReadyGood) {
		rc = check.OK
	} else if response.StatusCode != 200 {
		rc = check.Critical
	} else {
		rc = check.Unknown
	}

	return rc, output, err
}
