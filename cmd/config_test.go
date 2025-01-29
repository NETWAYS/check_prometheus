package cmd

import (
	"os"
	"testing"
)

func loadTestdata(filepath string) []byte {
	data, _ := os.ReadFile(filepath)
	return data
}

func TestConfig(t *testing.T) {
	c := cliConfig.NewClient()
	expected := "http://localhost:9090/"
	if c.URL != "http://localhost:9090/" {
		t.Error("\nActual: ", c.URL, "\nExpected: ", expected)
	}
}
