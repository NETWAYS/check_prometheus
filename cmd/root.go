package cmd

import (
	"github.com/NETWAYS/go-check"
	"github.com/spf13/cobra"
	"os"
)

const Copyright = `
Copyright (C) 2022 NETWAYS GmbH <info@netways.de>
`

const License = `
Copyright (C) 2022 NETWAYS GmbH <info@netways.de>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see https://www.gnu.org/licenses/.
`

var Timeout = 30

var rootCmd = &cobra.Command{
	Use:   "check_prometheus",
	Short: "An Icinga check plugin to check Prometheus",
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
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
		"Hostname of the Prometheus server")
	pfs.IntVarP(&cliConfig.Port, "port", "p", 9090,
		"Port of the Prometheus server")
	pfs.BoolVarP(&cliConfig.Secure, "secure", "s", false,
		"Use a HTTPS connection")
	pfs.BoolVarP(&cliConfig.Insecure, "insecure", "i", false,
		"Skip the verification of the server's TLS certificate")
	pfs.StringVarP(&cliConfig.Bearer, "bearer", "b", "",
		"Specify the Bearer Token for server authentication")
	pfs.StringVarP(&cliConfig.BasicAuth, "user", "u", "",
		"Specify the user name and password for server authentication <user:password>")
	pfs.StringVarP(&cliConfig.CAFile, "ca-file", "", "",
		"Specify the CA File for TLS authentication")
	pfs.StringVarP(&cliConfig.CertFile, "cert-file", "", "",
		"Specify the Certificate File for TLS authentication")
	pfs.StringVarP(&cliConfig.KeyFile, "key-file", "", "",
		"Specify the Key File for TLS authentication")
	pfs.IntVarP(&Timeout, "timeout", "t", Timeout,
		"Timeout in seconds for the CheckPlugin")

	rootCmd.Flags().SortFlags = false
	pfs.SortFlags = false

	help := rootCmd.HelpTemplate()
	rootCmd.SetHelpTemplate(help + Copyright)
}

func Usage(cmd *cobra.Command, strings []string) {
	_ = cmd.Usage()

	os.Exit(3)
}
