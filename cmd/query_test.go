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
	expected := "[UNKNOWN] - Post \"http://localhost:9999/api/v1/query\""

	if !strings.Contains(actual, expected) {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}

func TestQuery_MissingParameter(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "query")
	out, _ := cmd.CombinedOutput()

	expected := "[UNKNOWN] - required flag(s) \"query\" not set (*errors.errorString)"

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
			expected: "[UNKNOWN] - 0 Metrics: 0 Critical - 0 Warning - 0 Ok\nHTTP Warnings: hic sunt dracones, foo\n",
		},
		{
			name: "query-no-such-metric",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "foo"},
			expected: "[UNKNOWN] - 0 Metrics: 0 Critical - 0 Warning - 0 Ok\n | \nexit status 3\n",
		},
		{
			name: "query-no-such-matrix",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "foo"},
			expected: "[UNKNOWN] - 0 Metrics: 0 Critical - 0 Warning - 0 Ok\n | \nexit status 3\n",
		},
		{
			name: "query-scalar",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"scalar","result":[1670339013.992,"1"]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "1"},
			expected: "[UNKNOWN] - scalar value results are not supported (*errors.errorString)\nexit status 3\n",
		},
		{
			name: "query-string",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"string","result":[1670339013.992,"up"]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up"},
			expected: "[UNKNOWN] - string value results are not supported (*errors.errorString)\nexit status 3\n",
		},
		{
			name: "query-matrix-exists",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up","instance":"localhost","job":"node"},"values":[[1670340712.988,"1"],[1670340772.988,"1"],[1670340832.990,"1"],[1670340892.990,"1"],[1670340952.990,"1"]]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up{job=\"prometheus\"}[5m]"},
			expected: "[OK] - 1 Metrics OK | up_instance_localhost_job_node=1;10;20\n",
		},
		{
			name: "query-metric-exists",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","instance":"localhost","job":"prometheus"},"value":[1668782473.835,"1"]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up{job=\"prometheus\"}"},
			expected: "[OK] - 1 Metrics OK | up_instance_localhost_job_prometheus=1;10;20\n",
		},
		{
			name: "query-threshold-ok",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","instance":"localhost","job":"prometheus"},"value":[1668782473.835,"100"]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up{job=\"prometheus\"}", "-w", "0:", "-c", "0:"},
			expected: "[OK] - 1 Metrics OK | up_instance_localhost_job_prometheus=100\n",
		},
		{
			name: "query-threshold-critical",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","instance":"localhost","job":"prometheus"},"value":[1668782473.835,"-100"]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up{job=\"prometheus\"}", "-w", "0:", "-c", "0:"},
			expected: "[CRITICAL] - 1 Metrics: 1 Critical - 0 Warning - 0 Ok\n \\_[CRITICAL] up{instance=\"localhost\", job=\"prometheus\"} - value: -100\n | up_instance_localhost_job_prometheus=-100\nexit status 2\n",
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

func TestExtendedQueryCmd(t *testing.T) {
	tests := []QueryTest{
		{
			name: "vector-multiple-ok",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","instance":"localhost:9100","job":"node"},"value":[1696589905.608,"1"]},{"metric":{"__name__":"up","instance":"localhost:9104","job":"mysqld"},"value":[1696589905.608,"99"]},{"metric":{"__name__":"up","instance":"localhost:9117","job":"apache"},"value":[1696589905.608,"1"]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up", "-w", "100", "-c", "200"},
			expected: "[OK] - 3 Metrics OK |",
		},
		{
			name: "vector-multiple-critical",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","instance":"localhost:9100","job":"node"},"value":[1696589905.608,"1"]},{"metric":{"__name__":"up","instance":"localhost:9104","job":"mysqld"},"value":[1696589905.608,"11"]},{"metric":{"__name__":"up","instance":"localhost:9117","job":"apache"},"value":[1696589905.608,"6"]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up", "-w", "5", "-c", "10"},
			expected: "[CRITICAL] - 3 Metrics: 1 Critical - 1 Warning - 1 Ok",
		},
		{
			name: "matrix-multiple-critical",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up","instance":"localhost:9100","job":"node"},"values":[[1696589212.987,"1"],[1696589272.987,"1"],[1696589332.987,"1"],[1696589392.987,"1"],[1696589452.987,"1"]]},{"metric":{"__name__":"up","instance":"localhost:9104","job":"mysqld"},"values":[[1696589209.089,"25"],[1696589269.089,"25"],[1696589329.089,"25"],[1696589389.089,"25"],[1696589449.089,"25"]]},{"metric":{"__name__":"up","instance":"localhost:9117","job":"apache"},"values":[[1696589209.369,"1"],[1696589269.369,"1"],[1696589329.369,"1"],[1696589389.369,"1"],[1696589449.369,"1"]]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up", "-w", "10", "-c", "20"},
			expected: "[CRITICAL] - 3 Metrics: 1 Critical - 0 Warning - 2 Ok",
		},
		{
			name: "matrix-multiple-warning",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up","instance":"localhost:9100","job":"node"},"values":[[1696589212.987,"1"],[1696589272.987,"1"],[1696589332.987,"1"],[1696589392.987,"1"],[1696589452.987,"1"]]},{"metric":{"__name__":"up","instance":"localhost:9104","job":"mysqld"},"values":[[1696589209.089,"15"],[1696589269.089,"15"],[1696589329.089,"15"],[1696589389.089,"15"],[1696589449.089,"15"]]},{"metric":{"__name__":"up","instance":"localhost:9117","job":"apache"},"values":[[1696589209.369,"1"],[1696589269.369,"1"],[1696589329.369,"1"],[1696589389.369,"1"],[1696589449.369,"1"]]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up", "-w", "10", "-c", "20"},
			expected: "[WARNING] - 3 Metrics: 0 Critical - 1 Warning - 2 Ok",
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

			if !strings.Contains(actual, test.expected) {
				t.Error("\nActual: ", actual, "\nExpected: ", test.expected)
			}

		})
	}
}
