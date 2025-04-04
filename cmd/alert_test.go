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

	tests := []AlertTest{
		{
			name: "alert-none",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[]}}`))
			})),
			args:     []string{"run", "../main.go", "alert"},
			expected: "[OK] - No alerts defined | total=0 firing=0 pending=0 inactive=0\n",
		},
		{
			name: "alert-none-with-problems",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[]}}`))
			})),
			args:     []string{"run", "../main.go", "alert", "--problems"},
			expected: "[OK] - No alerts defined | total=0 firing=0 pending=0 inactive=0\n",
		},
		{
			name: "alert-none-with-no-state",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[]}}`))
			})),
			args:     []string{"run", "../main.go", "alert", "--no-alerts-state", "3"},
			expected: "[UNKNOWN] - No alerts defined | total=0 firing=0 pending=0 inactive=0\nexit status 3\n",
		},
		{
			name: "alert-none-with-name",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"groups":[]}}`))
			})),
			args:     []string{"run", "../main.go", "alert", "--name", "MyPreciousAlert"},
			expected: "[UNKNOWN] - No such alert defined | total=0 firing=0 pending=0 inactive=0\nexit status 3\n",
		},
		{
			name: "alert-default",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet1))
			})),
			args: []string{"run", "../main.go", "alert"},
			expected: `[CRITICAL] - 3 Alerts: 1 Firing - 1 Pending - 1 Inactive
\_ [OK] [HostOutOfMemory] is inactive
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00
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
			expected: `[CRITICAL] - 2 Alerts: 1 Firing - 1 Pending - 0 Inactive
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00
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
			expected: `[CRITICAL] - 1 Alerts: 1 Firing - 0 Pending - 0 Inactive
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00
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
			expected: `[CRITICAL] - 2 Alerts: 1 Firing - 1 Pending - 0 Inactive
\_ [WARNING] [SqlAccessDeniedRate] - Job: [mysql] on Instance: [localhost] is pending - value: 0.40
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00
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
			expected: `[CRITICAL] - 1 Alerts: 1 Firing - 0 Pending - 0 Inactive
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00
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
			args: []string{"run", "../main.go", "alert", "--name", "NoSuchAlert"},
			expected: `[UNKNOWN] - 0 Alerts: 0 Firing - 0 Pending - 0 Inactive
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
			args: []string{"run", "../main.go", "alert", "--name", "InactiveAlert", "--problems"},
			expected: `[UNKNOWN] - 0 Alerts: 0 Firing - 0 Pending - 0 Inactive
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
			expected: `[CRITICAL] - 2 Alerts: 1 Firing - 0 Pending - 1 Inactive
\_ [OK] [HostOutOfMemory] is inactive
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00
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
			expected: `[CRITICAL] - 1 Alerts: 1 Firing - 0 Pending - 0 Inactive
\_ [CRITICAL] [BlackboxTLS] - Job: [blackbox] on Instance: [https://localhost:443] is firing - value: -6065338.00
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
			expected: "[OK] - 1 Alerts: 0 Firing - 0 Pending - 1 Inactive\n\\_ [OK] [InactiveAlert] is inactive\n|total=1 firing=0 pending=0 inactive=1\n\n",
		},
		{
			name: "alert-recording-rule",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(loadTestdata(alertTestDataSet4))
			})),
			args:     []string{"run", "../main.go", "alert", "--name", "InactiveAlert"},
			expected: "[OK] - 1 Alerts: 0 Firing - 0 Pending - 1 Inactive\n\\_ [OK] [InactiveAlert] is inactive\n|total=1 firing=0 pending=0 inactive=1\n\n",
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
