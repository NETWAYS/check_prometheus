package client

import (
	"fmt"
	"net/http"
)

func newHttpRequest(requestURL string) (response *http.Response, err error) {
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	response, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) Health() (response *http.Response, err error) {
	return newHttpRequest(fmt.Sprintf("%s/-/healthy", c.Url))
}

func (c *Client) Ready() (response *http.Response, err error) {
	return newHttpRequest(fmt.Sprintf("%s/-/ready", c.Url))
}
