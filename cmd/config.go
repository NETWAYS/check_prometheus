package cmd

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NETWAYS/check_prometheus/internal/client"
	"github.com/NETWAYS/go-check"
	"github.com/prometheus/common/config"
)

type Config struct {
	BasicAuth string `env:"CHECK_PROMETHEUS_BASICAUTH"`
	Bearer    string `env:"CHECK_PROMETHEUS_BEARER"`
	CAFile    string `env:"CHECK_PROMETHEUS_CA_FILE"`
	CertFile  string `env:"CHECK_PROMETHEUS_CERT_FILE"`
	KeyFile   string `env:"CHECK_PROMETHEUS_KEY_FILE"`
	Hostname  string `env:"CHECK_PROMETHEUS_HOSTNAME"`
	URL       string `env:"CHECK_PROMETHEUS_URL"`
	Headers   []string
	Port      int
	Info      bool
	Insecure  bool
	PReady    bool
	Secure    bool
}

const Copyright = `
Copyright (C) 2022 NETWAYS GmbH <info@netways.de>
`

const License = `
Copyright (C) 2022 NETWAYS GmbH <info@netways.de>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see https://www.gnu.org/licenses/.
`

var cliConfig Config

func (c *Config) NewClient() *client.Client {
	u := url.URL{
		Scheme: "http",
		Host:   c.Hostname + ":" + strconv.Itoa(c.Port),
		Path:   c.URL,
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
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsConfig,
	}

	// Using a Bearer Token for authentication
	if c.Bearer != "" {
		var t = config.NewInlineSecret(c.Bearer)
		rt = config.NewAuthorizationCredentialsRoundTripper("Bearer", t, rt)
	}

	// Using a BasicAuth for authentication
	if c.BasicAuth != "" {
		s := strings.Split(c.BasicAuth, ":")
		if len(s) != 2 {
			check.ExitError(errors.New("specify the user name and password for server authentication <user:password>"))
		}

		var u = config.NewInlineSecret(s[0])

		var p = config.NewInlineSecret(s[1])

		rt = config.NewBasicAuthRoundTripper(u, p, rt)
	}

	// If extra headers are set, parse them and add them to the request
	if len(c.Headers) > 0 {
		headers := make(map[string]string)

		for _, h := range c.Headers {
			head := strings.Split(h, ":")
			if len(head) == 2 {
				headers[strings.TrimSpace(head[0])] = strings.TrimSpace(head[1])
			}
		}

		rt = client.NewHeadersRoundTripper(headers, rt)
	}

	return client.NewClient(u.String(), rt)
}

func (c *Config) timeoutContext() (context.Context, func()) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
