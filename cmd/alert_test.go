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
			name: "alert-default",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[{"name":"Foo","file":"alerts.yaml","rules":[{"state":"inactive","name":"HostOutOfMemory","query":"up","duration":120,"labels":{"severity":"critical"},"annotations":{"description":"Foo","summary":"Foo"},"alerts":[],"health":"ok","evaluationTime":0.000553928,"lastEvaluation":"2022-11-24T14:08:17.597083058Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000581212,"lastEvaluation":"2022-11-24T14:08:17.59706083Z"},{"name":"SQL","file":"alerts.yaml","rules":[{"state":"pending","name":"SqlAccessDeniedRate","query":"mysql","duration":17280000,"labels":{"severity":"warning"},"annotations":{"description":"MySQL","summary":"MySQL"},"alerts":[{"labels":{"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning"},"annotations":{"description":"MySQL","summary":"MySQL"},"state":"pending","activeAt":"2022-11-21T10:38:35.373483748Z","value":"4.03448275862069e-01"}],"health":"ok","evaluationTime":0.002909617,"lastEvaluation":"2022-11-24T14:08:25.375220595Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.003046259,"lastEvaluation":"2022-11-24T14:08:25.375096825Z"},{"name":"TLS","file":"alerts.yaml","rules":[{"state":"firing","name":"BlackboxTLS","query":"SSL","duration":0,"labels":{"severity":"critical"},"annotations":{"description":"TLS","summary":"TLS"},"alerts":[{"labels":{"alertname":"TLS","instance":"https://localhost:443","job":"blackbox","severity":"critical"},"annotations":{"description":"TLS","summary":"TLS"},"state":"firing","activeAt":"2022-11-24T05:11:27.211699259Z","value":"-6.065338210999966e+06"}],"health":"ok","evaluationTime":0.000713955,"lastEvaluation":"2022-11-24T14:08:17.212720815Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000738927,"lastEvaluation":"2022-11-24T14:08:17.212700182Z"}]}}`))
			})),
			args: []string{"run", "../main.go", "alert"},
			expected: `CRITICAL - 3 Alerts: 1 Firing - 1 Pending - 1 Inactive
 \_[OK] [HostOutOfMemory] is inactive
 \_[WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40
 \_[CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00
 | total=3 firing=1 pending=1 inactive=1
exit status 2
`,
		},
		{
			name: "alert-problems-only",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[{"name":"Foo","file":"alerts.yaml","rules":[{"state":"inactive","name":"HostOutOfMemory","query":"up","duration":120,"labels":{"severity":"critical"},"annotations":{"description":"Foo","summary":"Foo"},"alerts":[],"health":"ok","evaluationTime":0.000553928,"lastEvaluation":"2022-11-24T14:08:17.597083058Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000581212,"lastEvaluation":"2022-11-24T14:08:17.59706083Z"},{"name":"SQL","file":"alerts.yaml","rules":[{"state":"pending","name":"SqlAccessDeniedRate","query":"mysql","duration":17280000,"labels":{"severity":"warning"},"annotations":{"description":"MySQL","summary":"MySQL"},"alerts":[{"labels":{"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning"},"annotations":{"description":"MySQL","summary":"MySQL"},"state":"pending","activeAt":"2022-11-21T10:38:35.373483748Z","value":"4.03448275862069e-01"}],"health":"ok","evaluationTime":0.002909617,"lastEvaluation":"2022-11-24T14:08:25.375220595Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.003046259,"lastEvaluation":"2022-11-24T14:08:25.375096825Z"},{"name":"TLS","file":"alerts.yaml","rules":[{"state":"firing","name":"BlackboxTLS","query":"SSL","duration":0,"labels":{"severity":"critical"},"annotations":{"description":"TLS","summary":"TLS"},"alerts":[{"labels":{"alertname":"TLS","instance":"https://localhost:443","job":"blackbox","severity":"critical"},"annotations":{"description":"TLS","summary":"TLS"},"state":"firing","activeAt":"2022-11-24T05:11:27.211699259Z","value":"-6.065338210999966e+06"}],"health":"ok","evaluationTime":0.000713955,"lastEvaluation":"2022-11-24T14:08:17.212720815Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000738927,"lastEvaluation":"2022-11-24T14:08:17.212700182Z"}]}}`))
			})),
			args: []string{"run", "../main.go", "alert", "--problems"},
			expected: `CRITICAL - 2 Alerts: 1 Firing - 1 Pending - 0 Inactive
 \_[WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40
 \_[CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00
 | total=2 firing=1 pending=1 inactive=0
exit status 2
`,
		},
		{
			name: "alert-no-such-alert",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[{"name":"Foo","file":"alerts.yaml","rules":[{"state":"inactive","name":"InactiveAlert","query":"foo","duration":120,"labels":{"severity":"critical"},"annotations":{"description":"Inactive","summary":"Inactive"},"alerts":[],"health":"ok","evaluationTime":0.000462382,"lastEvaluation":"2022-11-18T14:01:07.597034323Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000478395,"lastEvaluation":"2022-11-18T14:01:07.597021953Z"}]}}`))
			})),
			args:     []string{"run", "../main.go", "alert", "--name", "NoSuchAlert"},
			expected: "UNKNOWN - 0 Alerts: 0 Firing - 0 Pending - 0 Inactive\n | \nexit status 3\n",
		},
		{
			name: "alert-inactive-with-problems",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[{"name":"Foo","file":"alerts.yaml","rules":[{"state":"inactive","name":"InactiveAlert","query":"foo","duration":120,"labels":{"severity":"critical"},"annotations":{"description":"Inactive","summary":"Inactive"},"alerts":[],"health":"ok","evaluationTime":0.000462382,"lastEvaluation":"2022-11-18T14:01:07.597034323Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000478395,"lastEvaluation":"2022-11-18T14:01:07.597021953Z"}]}}`))
			})),
			args:     []string{"run", "../main.go", "alert", "--name", "InactiveAlert", "--problems"},
			expected: "UNKNOWN - 0 Alerts: 0 Firing - 0 Pending - 0 Inactive\n | \nexit status 3\n",
		},
		{
			name: "alert-multiple-alerts",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[{"name":"Foo","file":"alerts.yaml","rules":[{"state":"inactive","name":"HostOutOfMemory","query":"up","duration":120,"labels":{"severity":"critical"},"annotations":{"description":"Foo","summary":"Foo"},"alerts":[],"health":"ok","evaluationTime":0.000553928,"lastEvaluation":"2022-11-24T14:08:17.597083058Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000581212,"lastEvaluation":"2022-11-24T14:08:17.59706083Z"},{"name":"SQL","file":"alerts.yaml","rules":[{"state":"pending","name":"SqlAccessDeniedRate","query":"mysql","duration":17280000,"labels":{"severity":"warning"},"annotations":{"description":"MySQL","summary":"MySQL"},"alerts":[{"labels":{"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning"},"annotations":{"description":"MySQL","summary":"MySQL"},"state":"pending","activeAt":"2022-11-21T10:38:35.373483748Z","value":"4.03448275862069e-01"}],"health":"ok","evaluationTime":0.002909617,"lastEvaluation":"2022-11-24T14:08:25.375220595Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.003046259,"lastEvaluation":"2022-11-24T14:08:25.375096825Z"},{"name":"TLS","file":"alerts.yaml","rules":[{"state":"firing","name":"BlackboxTLS","query":"SSL","duration":0,"labels":{"severity":"critical"},"annotations":{"description":"TLS","summary":"TLS"},"alerts":[{"labels":{"alertname":"TLS","instance":"https://localhost:443","job":"blackbox","severity":"critical"},"annotations":{"description":"TLS","summary":"TLS"},"state":"firing","activeAt":"2022-11-24T05:11:27.211699259Z","value":"-6.065338210999966e+06"}],"health":"ok","evaluationTime":0.000713955,"lastEvaluation":"2022-11-24T14:08:17.212720815Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000738927,"lastEvaluation":"2022-11-24T14:08:17.212700182Z"}]}}`))
			})),
			args: []string{"run", "../main.go", "alert", "--name", "HostOutOfMemory", "--name", "BlackboxTLS"},
			expected: `CRITICAL - 2 Alerts: 1 Firing - 0 Pending - 1 Inactive
 \_[OK] [HostOutOfMemory] is inactive
 \_[CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00
 | total=2 firing=1 pending=0 inactive=1
exit status 2
`,
		},
		{
			name: "alert-multiple-alerts-problems-only",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[{"name":"Foo","file":"alerts.yaml","rules":[{"state":"inactive","name":"HostOutOfMemory","query":"up","duration":120,"labels":{"severity":"critical"},"annotations":{"description":"Foo","summary":"Foo"},"alerts":[],"health":"ok","evaluationTime":0.000553928,"lastEvaluation":"2022-11-24T14:08:17.597083058Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000581212,"lastEvaluation":"2022-11-24T14:08:17.59706083Z"},{"name":"SQL","file":"alerts.yaml","rules":[{"state":"pending","name":"SqlAccessDeniedRate","query":"mysql","duration":17280000,"labels":{"severity":"warning"},"annotations":{"description":"MySQL","summary":"MySQL"},"alerts":[{"labels":{"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning"},"annotations":{"description":"MySQL","summary":"MySQL"},"state":"pending","activeAt":"2022-11-21T10:38:35.373483748Z","value":"4.03448275862069e-01"}],"health":"ok","evaluationTime":0.002909617,"lastEvaluation":"2022-11-24T14:08:25.375220595Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.003046259,"lastEvaluation":"2022-11-24T14:08:25.375096825Z"},{"name":"TLS","file":"alerts.yaml","rules":[{"state":"firing","name":"BlackboxTLS","query":"SSL","duration":0,"labels":{"severity":"critical"},"annotations":{"description":"TLS","summary":"TLS"},"alerts":[{"labels":{"alertname":"TLS","instance":"https://localhost:443","job":"blackbox","severity":"critical"},"annotations":{"description":"TLS","summary":"TLS"},"state":"firing","activeAt":"2022-11-24T05:11:27.211699259Z","value":"-6.065338210999966e+06"}],"health":"ok","evaluationTime":0.000713955,"lastEvaluation":"2022-11-24T14:08:17.212720815Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000738927,"lastEvaluation":"2022-11-24T14:08:17.212700182Z"}]}}`))
			})),
			args: []string{"run", "../main.go", "alert", "--name", "HostOutOfMemory", "--name", "BlackboxTLS", "--problems"},
			expected: `CRITICAL - 1 Alerts: 1 Firing - 0 Pending - 0 Inactive
 \_[CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00
 | total=1 firing=1 pending=0 inactive=0
exit status 2
`,
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
		{
			name: "alert-recording-rule",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[{"name":"example","file":"recoding.yaml","rules":[{"name":"job:foo","query":"sum by(job) (requests_total)","health":"ok","evaluationTime":0.000391321,"lastEvaluation":"2023-01-13T14:26:08.687065894Z","type":"recording"}],"interval":10,"evaluationTime":0.000403777,"lastEvaluation":"2023-01-13T14:26:08.687058029Z"},{"name":"Foo","file":"alerts.yaml","rules":[{"state":"inactive","name":"InactiveAlert","query":"foo","duration":120,"labels":{"severity":"critical"},"annotations":{"description":"Inactive","summary":"Inactive"},"alerts":[],"health":"ok","evaluationTime":0.000462382,"lastEvaluation":"2022-11-18T14:01:07.597034323Z","type":"alerting"}],"interval":10,"limit":0,"evaluationTime":0.000478395,"lastEvaluation":"2022-11-18T14:01:07.597021953Z"}]}}`))
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
