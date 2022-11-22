package cmd

import (
	"os/exec"
	"strings"
	"testing"
)

func TestQuery_ConnectionRefused(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "query", "--query", "foo", "--critical", "1", "--warning", "2", "--port", "9999")
	out, _ := cmd.CombinedOutput()

	expected := "UNKNOWN - Post \"http://localhost:9999/api/v1/query\": dial tcp 127.0.0.1:9999: connect: connection refused (*url.Error)\nexit status 3\n"

	actual := string(out)

	if actual != expected {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}

func TestQuery_MissingParameter(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "query")
	out, _ := cmd.CombinedOutput()

	expected := "UNKNOWN - please specify warning and critical thresholds (*errors.errorString)"

	actual := string(out)

	if !strings.Contains(actual, expected) {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}
