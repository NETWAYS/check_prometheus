package cmd

import (
	"context"
	"github.com/NETWAYS/check_prometheus/internal/client"
	"net/url"
	"strconv"
	"time"
)

type Config struct {
	Hostname string
	Port     int
	TLS      bool
	Insecure bool
}

var cliConfig Config

func (c *Config) Client() *client.Client {
	u := url.URL{
		Scheme: "http",
		Host:   c.Hostname + ":" + strconv.Itoa(c.Port),
	}

	if c.TLS {
		u.Scheme = "https"
	}

	cl := client.NewClient(u.String())
	cl.Insecure = c.Insecure

	return cl
}

func (c *Config) timeoutContext() (context.Context, func()) {
	// TODO Add timeout config
	return context.WithTimeout(context.Background(), 5*time.Second)
}
