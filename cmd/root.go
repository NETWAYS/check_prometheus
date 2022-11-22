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
		"Address of the prometheus instance")
	pfs.IntVarP(&cliConfig.Port, "port", "p", 9090,
		"Port of the prometheus instance")
	pfs.BoolVarP(&cliConfig.Secure, "secure", "s", false,
		"Use secure connection")
	pfs.BoolVarP(&cliConfig.Insecure, "insecure", "i", false,
		"Allow use of self signed certificates when using SSL")
	pfs.IntVarP(&Timeout, "timeout", "t", Timeout,
		"Timeout for the check")

	rootCmd.Flags().SortFlags = false
	pfs.SortFlags = false

	help := rootCmd.HelpTemplate()
	rootCmd.SetHelpTemplate(help + Copyright)
}

func Usage(cmd *cobra.Command, strings []string) {
	_ = cmd.Usage()

	os.Exit(3)
}
