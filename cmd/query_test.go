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
			name: "query-warning",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]},"warnings": ["hic sunt dracones", "foo"]}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "foo"},
			expected: "UNKNOWN - 0 Metrics: 0 Critical - 0 Warning - 0 Ok\nHTTP Warnings: hic sunt dracones, foo\n | \nexit status 3\n",
		},
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
			name: "query-no-such-matrix",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "foo"},
			expected: "UNKNOWN - 0 Metrics: 0 Critical - 0 Warning - 0 Ok\n | \nexit status 3\n",
		},
		{
			name: "query-scalar",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"scalar","result":[1670339013.992,"1"]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "1"},
			expected: "UNKNOWN - Scalar value results are not supported (*errors.errorString)\nexit status 3\n",
		},
		{
			name: "query-string",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"string","result":[1670339013.992,"up"]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up"},
			expected: "UNKNOWN - String value results are not supported (*errors.errorString)\nexit status 3\n",
		},
		{
			name: "query-matrix-exists",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up","instance":"localhost","job":"node"},"values":[[1670340712.988,"1"],[1670340772.988,"1"],[1670340832.990,"1"],[1670340892.990,"1"],[1670340952.990,"1"]]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up{job=\"prometheus\"}[5m]"},
			expected: "OK - 1 Metrics OK | up_instance_localhost_job_node=1\n",
		},
		{
			name: "query-metric-exists",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","instance":"localhost","job":"prometheus"},"value":[1668782473.835,"1"]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up{job=\"prometheus\"}"},
			expected: "OK - 1 Metrics OK | up_instance_localhost_job_prometheus=1\n",
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
