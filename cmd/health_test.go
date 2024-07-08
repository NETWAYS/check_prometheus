package cmd

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strings"
	"testing"
)

func TestHealth_ConnectionRefused(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "health", "--port", "9999")
	out, _ := cmd.CombinedOutput()

	actual := string(out)
	expected := "[UNKNOWN] - could not get status: Get \"http://localhost:9999/-/healthy\": dial"

	if !strings.Contains(actual, expected) {
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
			expected: "[OK] - states: ok=1\n\\_ [OK] Prometheus Server is Healthy.\n\n",
		},
		{
			name: "health-ok-older-versions",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`Prometheus is Healthy.`))
			})),
			args:     []string{"run", "../main.go", "health"},
			expected: "[OK] - states: ok=1\n\\_ [OK] Prometheus is Healthy.\n\n",
		},
		{
			name: "ready-ok",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`Prometheus Server is Ready.`))
			})),
			args:     []string{"run", "../main.go", "health", "--ready"},
			expected: "[OK] - states: ok=1\n\\_ [OK] Prometheus Server is Ready.\n\n",
		},
		{
			name: "info-ok",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"success","data":{"version":"2.30.3","revision":"foo","branch":"HEAD","buildUser":"root@foo","buildDate":"20211005-16:10:52","goVersion":"go1.17.1"}}`))
			})),
			args:     []string{"run", "../main.go", "health", "--info"},
			expected: "[OK] - states: ok=1\n\\_ [OK] Prometheus Server information\n\nVersion: 2.30.3\nBranch: HEAD\nBuildDate: 20211005-16:10:52\nBuildUser: root@foo\nRevision: foo\n\n",
		},
		{
			name: "health-bearer-ok",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				token := r.Header.Get("Authorization")
				if token == "Bearer secret" {
					// Just for testing, this is now how to handle tokens properly
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`Prometheus Server is Healthy.`))
					return
				}
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`The Authorization header wasn't set`))
			})),
			args:     []string{"run", "../main.go", "--bearer", "secret", "health"},
			expected: "[OK] - states: ok=1\n\\_ [OK] Prometheus Server is Healthy.\n\n",
		},
		{
			name: "health-bearer-unauthorized",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				token := r.Header.Get("Authorization")
				if token == "Bearer right-token" {
					// Just for testing, this is now how to handle BasicAuth properly
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`Prometheus Server is Healthy.`))
					return
				}
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`Access Denied!`))
			})),
			args:     []string{"run", "../main.go", "--bearer", "wrong-token", "health"},
			expected: "[CRITICAL] - states: critical=1\n\\_ [CRITICAL] Access Denied!\n\nexit status 2\n",
		},
		{
			name: "health-basic-auth-ok",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, pass, ok := r.BasicAuth()
				if ok {
					// Just for testing, this is now how to handle BasicAuth properly
					if user == "username" && pass == "password" {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`Prometheus Server is Healthy.`))
						return
					}
				}
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`The Authorization header wasn't set`))
			})),
			args:     []string{"run", "../main.go", "--user", "username:password", "health"},
			expected: "[OK] - states: ok=1\n\\_ [OK] Prometheus Server is Healthy.\n\n",
		},
		{
			name: "health-basic-auth-unauthorized",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user, pass, ok := r.BasicAuth()
				if ok {
					// Just for testing, this is now how to handle BasicAuth properly
					if user == "wrong" && pass == "kong" {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`Prometheus Server is Healthy.`))
						return
					}
				}
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`Access Denied!`))
			})),
			args:     []string{"run", "../main.go", "health"},
			expected: "[CRITICAL] - states: critical=1\n\\_ [CRITICAL] Access Denied!\n\nexit status 2\n",
		},
		{
			name: "health-basic-auth-wrong-use",
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`Access Denied!`))
			})),
			args:     []string{"run", "../main.go", "--user", "passwordmissing", "health"},
			expected: "[UNKNOWN] - specify the user name and password for server authentication <user:password> (*errors.errorString)\nexit status 3\n",
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
				t.Errorf("\nActual:\n%#v\nExpected:\n%#v\n", actual, test.expected)
			}

		})
	}
}
