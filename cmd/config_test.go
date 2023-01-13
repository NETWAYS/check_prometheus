package cmd

import (
	"testing"
)

func TestConfig(t *testing.T) {
	c := cliConfig.NewClient()
	expected := "http://localhost:9090"
	if c.Url != "http://localhost:9090" {
		t.Error("\nActual: ", c.Url, "\nExpected: ", expected)
	}
}
