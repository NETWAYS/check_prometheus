package cmd

import (
	"os/exec"
	"strings"
	"testing"
)

func TestQuery_ConnectionRefused(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "query", "--query", "foo", "--critical", "1", "--warning", "2", "--port", "9999")
	out, _ := cmd.CombinedOutput()

	actual := string(out)
	expected := "UNKNOWN - Post \"http://localhost:9999/api/v1/query\""

	if !strings.Contains(actual, expected) {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}

func TestQuery_MissingParameter(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "query")
	out, _ := cmd.CombinedOutput()

	expected := "UNKNOWN - required flag(s) \"query\" not set (*errors.errorString)"

	actual := string(out)

	if !strings.Contains(actual, expected) {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}
