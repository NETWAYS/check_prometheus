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

	queryTestDataSet1 := "../testdata/unittest/queryDataset1.json"

	queryTestDataSet2 := "../testdata/unittest/queryDataset2.json"

	queryTestDataSet3 := "../testdata/unittest/queryDataset3.json"

	queryTestDataSet4 := "../testdata/unittest/queryDataset4.json"

	tests := []QueryTest{
		{
			name: "query-warning",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]},"warnings": ["hic sunt dracones", "foo"]}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "foo"},
			expected: "[UNKNOWN] - No status information\nHTTP Warnings: hic sunt dracones, foo\n\nexit status 3\n",
		},
		{
			name: "query-no-such-metric",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "foo"},
			expected: "[UNKNOWN] - No status information\n\nexit status 3\n",
		},
		{
			name: "query-no-such-matrix",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "foo"},
			expected: "[UNKNOWN] - No status information\n\nexit status 3\n",
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
				w.Write(loadTestdata(queryTestDataSet1))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up{job=\"prometheus\"}[5m]"},
			expected: "[OK] - states: ok=1\n\\_ [OK]  1 @[1670340952.99] - value: 1\n|up_instance_localhost_job_node=1;10;20\n\n",
		},
		{
			name: "query-metric-exists2",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(queryTestDataSet2))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up{job=\"prometheus\"}"},
			expected: "[OK] - states: ok=1\n\\_ [OK]  up{instance=\"localhost\", job=\"prometheus\"} - value: 1\n|up_instance_localhost_job_prometheus=1;10;20\n\n",
		},
		{
			name: "query-threshold-ok",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(queryTestDataSet3))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up{job=\"prometheus\"}", "-w", "0:", "-c", "0:"},
			expected: "[OK] - states: ok=1\n\\_ [OK]  up{instance=\"localhost\", job=\"prometheus\"} - value: 100\n|up_instance_localhost_job_prometheus=100\n\n",
		},
		{
			name: "query-threshold-critical",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(queryTestDataSet4))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up{job=\"prometheus\"}", "-w", "0:", "-c", "0:"},
			expected: "[CRITICAL] - states: critical=1\n\\_ [CRITICAL]  up{instance=\"localhost\", job=\"prometheus\"} - value: -100\n|up_instance_localhost_job_prometheus=-100\n\nexit status 2\n",
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
				//				t.Error("\nActual: ", actual, "\nExpected: ", test.expected)
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
			expected: "OK] - states: ok=3\n\\_ [OK]  up{instance=\"localhost:9100\", job=\"node\"} - value: 1\n\\_ [OK]  up{instance=\"localhost:9104\", job=\"mysqld\"} - value: 99\n\\_ [OK]  up{instance=\"localhost:9117\", job=\"apache\"} - value: 1\n|up_instance_localhost:9100_job_node=1;100;200 up_instance_localhost:9104_job_mysqld=99;100;200 up_instance_localhost:9117_job_apache=1;100;200\n\n",
		},
		{
			name: "vector-multiple-critical",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up","instance":"localhost:9100","job":"node"},"value":[1696589905.608,"1"]},{"metric":{"__name__":"up","instance":"localhost:9104","job":"mysqld"},"value":[1696589905.608,"11"]},{"metric":{"__name__":"up","instance":"localhost:9117","job":"apache"},"value":[1696589905.608,"6"]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up", "-w", "5", "-c", "10"},
			expected: "[CRITICAL] - states: critical=1 warning=1 ok=1\n\\_ [OK]  up{instance=\"localhost:9100\", job=\"node\"} - value: 1\n\\_ [CRITICAL]  up{instance=\"localhost:9104\", job=\"mysqld\"} - value: 11\n\\_ [WARNING]  up{instance=\"localhost:9117\", job=\"apache\"} - value: 6\n|up_instance_localhost:9100_job_node=1;5;10 up_instance_localhost:9104_job_mysqld=11;5;10 up_instance_localhost:9117_job_apache=6;5;10\n\nexit status 2\n",
		},
		{
			name: "matrix-multiple-critical",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up","instance":"localhost:9100","job":"node"},"values":[[1696589212.987,"1"],[1696589272.987,"1"],[1696589332.987,"1"],[1696589392.987,"1"],[1696589452.987,"1"]]},{"metric":{"__name__":"up","instance":"localhost:9104","job":"mysqld"},"values":[[1696589209.089,"25"],[1696589269.089,"25"],[1696589329.089,"25"],[1696589389.089,"25"],[1696589449.089,"25"]]},{"metric":{"__name__":"up","instance":"localhost:9117","job":"apache"},"values":[[1696589209.369,"1"],[1696589269.369,"1"],[1696589329.369,"1"],[1696589389.369,"1"],[1696589449.369,"1"]]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up", "-w", "10", "-c", "20"},
			expected: "[CRITICAL] - states: critical=1 ok=2\n\\_ [OK]  1 @[1696589452.987] - value: 1\n\\_ [CRITICAL]  25 @[1696589449.089] - value: 25\n\\_ [OK]  1 @[1696589449.369] - value: 1\n|up_instance_localhost:9100_job_node=1;10;20 up_instance_localhost:9104_job_mysqld=25;10;20 up_instance_localhost:9117_job_apache=1;10;20\n\nexit status 2\n",
		},
		{
			name: "matrix-multiple-warning",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"__name__":"up","instance":"localhost:9100","job":"node"},"values":[[1696589212.987,"1"],[1696589272.987,"1"],[1696589332.987,"1"],[1696589392.987,"1"],[1696589452.987,"1"]]},{"metric":{"__name__":"up","instance":"localhost:9104","job":"mysqld"},"values":[[1696589209.089,"15"],[1696589269.089,"15"],[1696589329.089,"15"],[1696589389.089,"15"],[1696589449.089,"15"]]},{"metric":{"__name__":"up","instance":"localhost:9117","job":"apache"},"values":[[1696589209.369,"1"],[1696589269.369,"1"],[1696589329.369,"1"],[1696589389.369,"1"],[1696589449.369,"1"]]}]}}`))
			})),
			args:     []string{"run", "../main.go", "query", "--query", "up", "-w", "10", "-c", "20"},
			expected: "WARNING] - states: warning=1 ok=2\n\\_ [OK]  1 @[1696589452.987] - value: 1\n\\_ [WARNING]  15 @[1696589449.089] - value: 15\n\\_ [OK]  1 @[1696589449.369] - value: 1\n|up_instance_localhost:9100_job_node=1;10;20 up_instance_localhost:9104_job_mysqld=15;10;20 up_instance_localhost:9117_job_apache=1;10;20\n\nexit status 1\n",
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
