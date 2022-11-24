package cmd

import (
	"context"
	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Checks the health or readiness status of the Prometheus server",
	Long: `Checks the health or readiness status of the Prometheus server
Health: Checks the health of an endpoint, which returns OK if the Prometheus server is healthy.
Ready: Checks the readiness of an endpoint, which returns OK if the Prometheus server is ready to serve traffic (i.e. respond to queries).`,
	Example: `  $ check_prometheus health --hostname 'localhost' --port 9090 --insecure           
  OK - Prometheus Server is Healthy. | statuscode=200`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			rc     int
			output string
		)

		// Creating an client and connecting to the API
		c := cliConfig.Client()
		err := c.Connect()
		if err != nil {
			check.ExitError(err)
		}

		// Getting the preconfigured context
		ctx, cancel := cliConfig.timeoutContext()
		defer cancel()

		if cliConfig.PReady {
			// Getting the ready status
			rc, output, err = c.GetStatus(ctx, "ready")

			if err != nil {
				check.ExitError(err)
			}

		} else {
			// Getting the health status
			rc, output, err = c.GetStatus(ctx, "healthy")

			if err != nil {
				check.ExitError(err)
			}
		}

		if cliConfig.Info {
			// Displays various build information properties about the Prometheus server
			info, err := c.Api.Buildinfo(context.Background())
			if err != nil {
				check.ExitError(err)
			}

			output = "Prometheus Server information\n\n" +
				"Version: " + info.Version + "\n" +
				"Branch: " + info.Branch + "\n" +
				"BuildDate: " + info.BuildDate + "\n" +
				"BuildUser: " + info.BuildUser + "\n" +
				"Revision: " + info.Revision

			p := perfdata.PerfdataList{
				{Label: "status", Value: rc},
				{Label: "version", Value: info.Version},
				{Label: "builddate", Value: info.BuildDate},
				{Label: "builduser", Value: info.BuildUser},
				{Label: "revision", Value: info.Revision},
			}

			check.ExitRaw(rc, output, "|", p.String())
		}

		check.ExitRaw(rc, output)
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)

	fs := healthCmd.Flags()
	fs.BoolVarP(&cliConfig.PReady, "ready", "r", false,
		"Checks the readiness of an endpoint")
	fs.BoolVarP(&cliConfig.Info, "info", "I", false,
		"Displays various build information properties about the Prometheus server")

	fs.SortFlags = false
}
