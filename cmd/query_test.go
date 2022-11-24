package cmd

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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

type QueryTest struct {
	name     string
	server   *httptest.Server
	args     []string
	expected string
}

func TestQueryCmd(t *testing.T) {
	tests := []QueryTest{
		{
			name: "query-no-such-metric",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "foo"},
			expected: "UNKNOWN - 0 Metrics: 0 Critical - 0 Warning - 0 Ok\n | \nexit status 3\n",
		},
		{
			name: "query-metric-exists",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","instance":"localhost","job":"prometheus"},"value":[1668782473.835,"1"]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up{job=\"prometheus\"}"},
			expected: "OK - 1 Metrics OK | value_up_localhost_prometheus=1\n",
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
