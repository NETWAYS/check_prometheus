package cmd

import (
	"os/exec"
	"testing"
)

func TestAlert_ConnectionRefused(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "alert", "--port", "9999")
	out, _ := cmd.CombinedOutput()

	expected := "UNKNOWN - Get \"http://localhost:9999/api/v1/rules\": dial tcp 127.0.0.1:9999: connect: connection refused (*url.Error)\nexit status 3\n"

	actual := string(out)

	if actual != expected {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}
