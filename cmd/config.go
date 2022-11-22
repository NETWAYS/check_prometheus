package cmd

import (
	"context"
	"github.com/NETWAYS/check_prometheus/internal/client"
	"net/url"
	"strconv"
	"time"
)

type AlertConfig struct {
	AlertName []string
	Group     []string
	Problems  bool
}

type Config struct {
	Hostname string
	Port     int
	Secure   bool
	Insecure bool
	PReady   bool
	Info     bool
}

var (
	cliConfig      Config
	cliAlertConfig AlertConfig
)

func (c *Config) Client() *client.Client {
	u := url.URL{
		Scheme: "http",
		Host:   c.Hostname + ":" + strconv.Itoa(c.Port),
	}

	if c.Secure {
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
