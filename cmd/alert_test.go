package cmd

import (
	"os/exec"
	"strings"
	"testing"
)

func TestAlert_ConnectionRefused(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "alert", "--port", "9999")
	out, _ := cmd.CombinedOutput()

	actual := string(out)
	expected := "UNKNOWN - Get \"http://localhost:9999/api/v1/rules\""

	if !strings.Contains(actual, expected) {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}
