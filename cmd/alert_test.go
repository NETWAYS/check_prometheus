package cmd

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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

type AlertTest struct {
	name     string
	server   *httptest.Server
	args     []string
	expected string
}

func TestAlertCmd(t *testing.T) {
	tests := []AlertTest{
		{
			name: "alert-no-such-alert",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[{"name":"Foo","file":"alerts.yaml","rules":[{"state":"inactive","name":"InactiveAlert","query":"foo","duration":120,"labels":{"severity":"critical"},"annotations":{"description":"Inactive","summary":"Inactive"},"alerts":[],"health":"ok","evaluationTime":0.000462382,"lastEvaluation":"2022-11-18T14:01:07.597034323Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000478395,"lastEvaluation":"2022-11-18T14:01:07.597021953Z"}]}}`))
			})),
			args:     []string{"run", "../main.go", "alert", "--name", "NoSuchAlert"},
			expected: "UNKNOWN - 0 Alerts: 0 Firing - 0 Pending - 0 Inactive \n | \nexit status 3\n",
		},
		{
			name: "alert-inactive",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[{"name":"Foo","file":"alerts.yaml","rules":[{"state":"inactive","name":"InactiveAlert","query":"foo","duration":120,"labels":{"severity":"critical"},"annotations":{"description":"Inactive","summary":"Inactive"},"alerts":[],"health":"ok","evaluationTime":0.000462382,"lastEvaluation":"2022-11-18T14:01:07.597034323Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000478395,"lastEvaluation":"2022-11-18T14:01:07.597021953Z"}]}}`))
			})),
			args:     []string{"run", "../main.go", "alert", "--name", "InactiveAlert"},
			expected: "OK - Alerts inactive | firing=0 pending=0 inactive=1\n",
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
