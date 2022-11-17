package cmd

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"testing"
)

func TestHealth_ConnectionRefused(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "health", "--port", "9999")
	out, _ := cmd.CombinedOutput()

	expected := "UNKNOWN - Get \"http://localhost:9999/-/healthy\": dial tcp 127.0.0.1:9999: connect: connection refused (*url.Error)\nexit status 3\n"

	actual := string(out)

	if actual != expected {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}

type HealthTest struct {
	name     string
	server   *httptest.Server
	args     []string
	expected string
}

func TestHealthCmd(t *testing.T) {
	tests := []HealthTest{
		{
			name: "health-ok",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`Prometheus Server is Healthy.`))
			})),
			args:     []string{"run", "../main.go", "health"},
			expected: "OK - Prometheus Server is Healthy. | statuscode=200\n",
		},
		{
			name: "ready-ok",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`Prometheus Server is Ready.`))
			})),
			args:     []string{"run", "../main.go", "health", "--ready"},
			expected: "OK - Prometheus Server is Ready. | statuscode=200\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.server.Close()

			// We need the random Port extracted
			u, _ := url.Parse(test.server.URL)
			cmd := exec.Command("go", append(test.args, "--port", u.Port())...)
			out, _ := cmd.CombinedOutput()

			actual := string(out)

			if actual != test.expected {
				t.Error("\nActual: ", actual, "\nExpected: ", test.expected)
			}

		})
	}
}
