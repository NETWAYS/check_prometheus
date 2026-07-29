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
	expected := "[UNKNOWN] - Get \"http://localhost:9999/api/v1/rules\""

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

	alertTestDataSet1 := "../testdata/unittest/alertDataset1.json"

	alertTestDataSet2 := "../testdata/unittest/alertDataset2.json"

	alertTestDataSet3 := "../testdata/unittest/alertDataset3.json"

	alertTestDataSet4 := "../testdata/unittest/alertDataset4.json"

	alertTestDataSet5 := "../testdata/unittest/alertDataset5.json"

	tests := []AlertTest{
		{
			name: "alert-none",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[]}}`))
			})),
			args:     []string{"run", "../main.go", "alert"},
			expected: "[OK] - No alerts defined|total=0 firing=0 pending=0 inactive=0\n",
		},
		{
			name: "alert-none-with-problems",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[]}}`))
			})),
			args:     []string{"run", "../main.go", "alert", "--problems"},
			expected: "[OK] - No alerts defined|total=0 firing=0 pending=0 inactive=0\n",
		},
		{
			name: "alert-none-with-no-state",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[]}}`))
			})),
			args:     []string{"run", "../main.go", "alert", "--no-alerts-state", "3"},
			expected: "[UNKNOWN] - No alerts defined|total=0 firing=0 pending=0 inactive=0\nexit status 3\n",
		},
		{
			name: "alert-none-with-name",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[]}}`))
			})),
			args:     []string{"run", "../main.go", "alert", "--name", "MyPreciousAlert"},
			expected: "[UNKNOWN] - No such alert defined|total=0 firing=0 pending=0 inactive=0\nexit status 3\n",
		},
		{
			name: "alert-default",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert"},
			expected: `[CRITICAL] - [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
\_ [OK] [HostOutOfMemory] is inactive
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=3 firing=1 pending=1 inactive=1
exit status 2
`,
		},
		{
			name: "alert-problems-only",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--problems"},
			expected: `[CRITICAL] - [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=2 firing=1 pending=1 inactive=0
exit status 2
`,
		},
		{
			name: "alert-problems-only-with-exlude-on-one-group",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--problems", "-g", "TLS"},
			expected: `[CRITICAL] - [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=1 firing=1 pending=0 inactive=0
exit status 2
`,
		},
		{
			name: "alert-problems-only-with-exlude-on-two-groups",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--problems", "-g", "SQL", "-g", "TLS"},
			expected: `[CRITICAL] - [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=2 firing=1 pending=1 inactive=0
exit status 2
`,
		},
		{
			name: "alert-problems-only-with-exlude",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--problems", "--exclude-alert", "Sql.*DeniedRate"},
			expected: `[CRITICAL] - [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=1 firing=1 pending=0 inactive=0
exit status 2
`,
		},
		{
			name: "alert-with-exclude-error",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet3))
			})),
			args:     []string{"run", "../main.go", "alert", "--exclude-alert", "[a-z"},
			expected: "[UNKNOWN] - Invalid regular expression provided: error parsing regexp: missing closing ]: `[a-z`\nexit status 3\n",
		},
		{
			name: "alert-no-such-alert",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet2))
			})),
			args: []string{"run", "../main.go", "alert", "--name", "NoSuchAlert", "-T", "3"},
			expected: `[UNKNOWN] - No alerts retrieved
\_ [UNKNOWN] No alerts retrieved
|total=0 firing=0 pending=0 inactive=0
exit status 3
`,
		},
		{
			name: "alert-inactive-with-problems",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet2))
			})),
			args: []string{"run", "../main.go", "alert", "--name", "InactiveAlert", "--problems", "-T", "3"},
			expected: `[UNKNOWN] - No alerts retrieved
\_ [UNKNOWN] No alerts retrieved
|total=0 firing=0 pending=0 inactive=0
exit status 3
`,
		},
		{
			name: "alert-multiple-alerts",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--name", "HostOutOfMemory", "--name", "BlackboxTLS"},
			expected: `[CRITICAL] - [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
\_ [OK] [HostOutOfMemory] is inactive
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=2 firing=1 pending=0 inactive=1
exit status 2
`,
		},
		{
			name: "alert-multiple-alerts-problems-only",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--name", "HostOutOfMemory", "--name", "BlackboxTLS", "--problems"},
			expected: `[CRITICAL] - [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=1 firing=1 pending=0 inactive=0
exit status 2
`,
		},
		{
			name: "alert-inactive",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet2))
			})),
			args:     []string{"run", "../main.go", "alert", "--name", "InactiveAlert"},
			expected: "[OK] - 1 Alerts: 0 Firing - 0 Pending - 1 Inactive\n\\_ [OK] [InactiveAlert] is inactive\n|total=1 firing=0 pending=0 inactive=1\n",
		},
		{
			name: "alert-watchdog",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet2))
			})),
			args:     []string{"run", "../main.go", "alert", "--name", "InactiveAlert", "-W"},
			expected: "[CRITICAL] - [InactiveAlert] is inactive\n\\_ [CRITICAL] [InactiveAlert] is inactive\n|total=1 firing=0 pending=0 inactive=1\nexit status 2\n",
		},
		{
			name: "alert-recording-rule",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet4))
			})),
			args:     []string{"run", "../main.go", "alert", "--name", "InactiveAlert"},
			expected: "[OK] - 1 Alerts: 0 Firing - 0 Pending - 1 Inactive\n\\_ [OK] [InactiveAlert] is inactive\n|total=1 firing=0 pending=0 inactive=1\n",
		},
		{
			name: "alert-include-label",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--include-label", "severity=critical"},
			expected: `[CRITICAL] - [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
\_ [OK] [HostOutOfMemory] is inactive
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=3 firing=1 pending=1 inactive=1
exit status 2
`,
		},
		{
			name: "alert-exclude-label",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--exclude-label", "severity=critical"},
			expected: `[WARNING] - [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
|total=1 firing=0 pending=1 inactive=0
exit status 1
`,
		},
		{
			name: "alert-exclude-label-regex",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--exclude-label", "severity=crit.*"},
			expected: `[WARNING] - [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
|total=1 firing=0 pending=1 inactive=0
exit status 1
`,
		},
		{
			name: "alert-include-label-multiple",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--include-label", "team=database", "--include-label", "severity=critical"},
			expected: `[CRITICAL] - [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
\_ [OK] [HostOutOfMemory] is inactive
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=3 firing=1 pending=1 inactive=1
exit status 2
`,
		},
		{
			name: "alert-include-label-multiple-regex",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--include-label", "team=data.+", "--include-label", "severity=critical"},
			expected: `[CRITICAL] - [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
\_ [OK] [HostOutOfMemory] is inactive
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=3 firing=1 pending=1 inactive=1
exit status 2
`,
		},
		{
			name: "alert-include-label-multiple-similar",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--include-label", "severity=warning", "--include-label", "severity=critical"},
			expected: `[CRITICAL] - [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
\_ [OK] [HostOutOfMemory] is inactive
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=3 firing=1 pending=1 inactive=1
exit status 2
`,
		},
		{
			name: "alert-exclude-label-multiple",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args:     []string{"run", "../main.go", "alert", "--exclude-label", "team=database", "--exclude-label", "severity=critical"},
			expected: "[OK] - 0 Alerts: 0 Firing - 0 Pending - 0 Inactive\n\\_ [OK] No alerts retrieved\n|total=0 firing=0 pending=0 inactive=0\n",
		},
		{
			name: "alert-state-label",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert", "--label-key-state=icinga"},
			expected: `[WARNING] - [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
\_ [OK] [HostOutOfMemory] is inactive
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40 - {"alertname":"SqlAccessDeniedRate","instance":"localhost","job":"mysql","severity":"warning","team":"database"}
\_ [OK] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00 - {"alertname":"TLS","icinga":"ok","instance":"https://localhost:443","job":"blackbox","severity":"critical","team":"network"}
|total=3 firing=1 pending=1 inactive=1
exit status 1
`,
		},
		{
			name: "alert-include-with-name-and-regex",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet5))
			})),
			args: []string{"run", "../main.go", "alert", "--name", "ContainerKilled", "--include-label", "name=(mosquitto|nodered)"},
			expected: `[CRITICAL] - [ContainerKilled] - Job: [cadvisor] on Instance: [example:8888] is firing - value: 123.40 - {"alertname":"ContainerKilled","instance":"example:8888","job":"cadvisor","name":"nodered","severity":"warning"}
\_ [CRITICAL] [ContainerKilled] - Job: [cadvisor] on Instance: [example:8888] is firing - value: 123.40 - {"alertname":"ContainerKilled","instance":"example:8888","job":"cadvisor","name":"nodered","severity":"warning"}
\_ [CRITICAL] [ContainerKilled] - Job: [cadvisor] on Instance: [example:8888] is firing - value: 123.40 - {"alertname":"ContainerKilled","instance":"example:8888","job":"cadvisor","name":"mosquitto","severity":"warning"}
|total=3 firing=3 pending=0 inactive=0
exit status 2
`,
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
