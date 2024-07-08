package cmd

import (
	"fmt"

	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/result"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Checks the health or readiness status of the Prometheus server",
	Long: `Checks the health or readiness status of the Prometheus server
Health: Checks the health of an endpoint, which returns OK if the Prometheus server is healthy.
Ready: Checks the readiness of an endpoint, which returns OK if the Prometheus server is ready to serve traffic (i.e. respond to queries).`,
	Example: `
	$ check_prometheus health --hostname 'localhost' --port 9090 --insecure
	OK - Prometheus Server is Healthy. | statuscode=200

	$ check_prometheus --bearer secrettoken health --ready
	OK - Prometheus Server is Ready. | statuscode=200`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			rc int
		)

		overall := result.Overall{}

		// Creating an client and connecting to the API
		c := cliConfig.NewClient()
		err := c.Connect()
		if err != nil {
			check.ExitError(err)
		}

		// Getting the preconfigured context
		ctx, cancel := cliConfig.timeoutContext()
		defer cancel()

		if cliConfig.PReady {
			// Getting the ready status
			rc, _, output, err := c.GetStatus(ctx, "ready")

			if err != nil {
				check.ExitError(fmt.Errorf(output))
			}

			partialResult := result.NewPartialResult()

			partialResult.SetState(rc)
			partialResult.Output = output
			overall.AddSubcheck(partialResult)

			check.ExitRaw(overall.GetStatus(), overall.GetOutput())
		}

		if cliConfig.Info {
			// Displays various build information properties about the Prometheus server
			info, err := c.API.Buildinfo(ctx)
			if err != nil {
				check.ExitError(err)
			}
			partialResult := result.NewPartialResult()

			partialResult.SetState(rc)

			partialResult.Output = "Prometheus Server information\n\n" +
				"Version: " + info.Version + "\n" +
				"Branch: " + info.Branch + "\n" +
				"BuildDate: " + info.BuildDate + "\n" +
				"BuildUser: " + info.BuildUser + "\n" +
				"Revision: " + info.Revision

			overall.AddSubcheck(partialResult)

			check.ExitRaw(overall.GetStatus(), overall.GetOutput())
		}

		// Getting the health status is the default
		rc, _, output, err := c.GetStatus(ctx, "healthy")

		if err != nil {
			check.ExitError(fmt.Errorf(output))
		}

<<<<<<< HEAD
		p := perfdata.PerfdataList{
			{Label: "statuscode", Value: float64(sc)},
		}
=======
		partialResult := result.NewPartialResult()
		partialResult.SetState(rc)
		partialResult.Output = output
		overall.AddSubcheck(partialResult)
>>>>>>> 9761a63 (wip: use more gocheck)

		check.ExitRaw(overall.GetStatus(), overall.GetOutput())
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
