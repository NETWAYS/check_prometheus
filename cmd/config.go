package cmd

import (
	"context"
	"fmt"
	"github.com/NETWAYS/check_prometheus/internal/client"
	"github.com/NETWAYS/go-check"
	"github.com/prometheus/common/config"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type AlertConfig struct {
	AlertName []string
	Group     []string
	Problems  bool
}

type Config struct {
	BasicAuth string
	Bearer    string
	CAFile    string
	CertFile  string
	KeyFile   string
	Hostname  string
	Info      bool
	Insecure  bool
	PReady    bool
	Port      int
	Secure    bool
}

var (
	cliConfig      Config
	cliAlertConfig AlertConfig
)

// TODO: Rename to NewClient
func (c *Config) Client() *client.Client {
	u := url.URL{
		Scheme: "http",
		Host:   c.Hostname + ":" + strconv.Itoa(c.Port),
	}

	if c.Secure {
		u.Scheme = "https"
	}

	// Create TLS configuration for default RoundTripper
	tlsConfig, err := config.NewTLSConfig(&config.TLSConfig{
		InsecureSkipVerify: c.Insecure,
		CAFile:             c.CAFile,
		KeyFile:            c.KeyFile,
		CertFile:           c.CertFile,
	})

	if err != nil {
		check.ExitError(err)
	}

	var rt http.RoundTripper = &http.Transport{
		TLSClientConfig:       tlsConfig,
		IdleConnTimeout:       10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	}

	// Using a Bearer Token for authentication
	if c.Bearer != "" {
		var t config.Secret = config.Secret(c.Bearer)
		rt = config.NewAuthorizationCredentialsRoundTripper("Bearer", t, rt)
	}

	// Using a BasicAuth for authentication
	if c.BasicAuth != "" {
		s := strings.Split(c.BasicAuth, ":")
		if len(s) != 2 {
			check.ExitError(fmt.Errorf("Specify the user name and password for server authentication <user:password>"))
		}
		var u string = s[0]
		var p config.Secret = config.Secret(s[1])
		rt = config.NewBasicAuthRoundTripper(u, p, "", rt)
	}

	return client.NewClient(u.String(), rt)
}

func (c *Config) timeoutContext() (context.Context, func()) {
	// TODO Add timeout config
	return context.WithTimeout(context.Background(), 5*time.Second)
}
