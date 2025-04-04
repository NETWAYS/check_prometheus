package cmd

import (
	"os"

	"github.com/NETWAYS/go-check"
	"github.com/spf13/cobra"
)

var Timeout = 30

var rootCmd = &cobra.Command{
	Use:   "check_prometheus",
	Short: "An Icinga check plugin to check Prometheus",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		go check.HandleTimeout(Timeout)
	},
	Run: Usage,
}

func Execute(version string) {
	defer check.CatchPanic()

	rootCmd.Version = version
	rootCmd.VersionTemplate()

	if err := rootCmd.Execute(); err != nil {
		check.ExitError(err)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.DisableAutoGenTag = true

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})

	pfs := rootCmd.PersistentFlags()
	pfs.StringVarP(&cliConfig.Hostname, "hostname", "H", "localhost",
		"Hostname of the Prometheus server (CHECK_PROMETHEUS_HOSTNAME)")
	pfs.IntVarP(&cliConfig.Port, "port", "p", 9090,
		"Port of the Prometheus server")
	pfs.StringVarP(&cliConfig.URL, "url", "U", "/",
		"URL/Path to append to the Promethes Hostname (CHECK_PROMETHEUS_URL)")
	pfs.BoolVarP(&cliConfig.Secure, "secure", "s", false,
		"Use a HTTPS connection")
	pfs.BoolVarP(&cliConfig.Insecure, "insecure", "i", false,
		"Skip the verification of the server's TLS certificate")
	pfs.StringVarP(&cliConfig.Bearer, "bearer", "b", "",
		"Specify the Bearer Token for server authentication (CHECK_PROMETHEUS_BEARER)")
	pfs.StringVarP(&cliConfig.BasicAuth, "user", "u", "",
		"Specify the user name and password for server authentication <user:password> (CHECK_PROMETHEUS_BASICAUTH)")
	pfs.StringVarP(&cliConfig.CAFile, "ca-file", "", "",
		"Specify the CA File for TLS authentication (CHECK_PROMETHEUS_CA_FILE)")
	pfs.StringVarP(&cliConfig.CertFile, "cert-file", "", "",
		"Specify the Certificate File for TLS authentication (CHECK_PROMETHEUS_CERT_FILE)")
	pfs.StringVarP(&cliConfig.KeyFile, "key-file", "", "",
		"Specify the Key File for TLS authentication (CHECK_PROMETHEUS_KEY_FILE)")
	pfs.IntVarP(&Timeout, "timeout", "t", Timeout,
		"Timeout in seconds for the CheckPlugin")
	pfs.StringSliceVarP(&cliConfig.Headers, "header", "", nil,
		`Additional HTTP header to include in the request. Can be used multiple times.
Keys and values are separated by a colon (--header "X-Custom: example").`)

	rootCmd.Flags().SortFlags = false
	pfs.SortFlags = false

	help := rootCmd.HelpTemplate()
	rootCmd.SetHelpTemplate(help + Copyright)

	check.LoadFromEnv(&cliConfig)
}

func Usage(cmd *cobra.Command, _ []string) {
	_ = cmd.Usage()

	os.Exit(3)
}
