package cmd

import (
	"context"
	"fmt"
	"github.com/NETWAYS/go-check"
	"github.com/spf13/cobra"
	"io/ioutil"
	"strconv"
	"strings"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Checks the health status of the Prometheus instance",
	Long:  `Checks the health status of the Prometheus instance. With `,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			rc     int
			output string
		)

		c := cliConfig.Client()
		err := c.Connect()
		if err != nil {
			check.ExitError(err)
		}

		// Health status
		health, err := c.Health()
		if err != nil {
			check.ExitError(err)
		}

		healthBody, err := ioutil.ReadAll(health.Body)
		if err != nil {
			check.ExitError(fmt.Errorf("could not read response body: %w", err))
		}

		out := strings.TrimSpace(string(healthBody))

		if health.StatusCode != 200 {
			rc = check.Unknown
		} else {
			rc = check.OK
		}

		output += strconv.Itoa(health.StatusCode) + ": " + out

		// Ready status
		if cliConfig.PReady {
			ready, err := c.Ready()
			if err != nil {
				check.ExitError(err)
			}

			readybody, err := ioutil.ReadAll(ready.Body)
			if err != nil {
				check.ExitError(err)
			}

			out := strings.TrimSpace(string(readybody))

			if ready.StatusCode != 200 {
				rc = check.Unknown
			} else {
				rc = check.OK
			}

			output += " - " + strconv.Itoa(ready.StatusCode) + ": " + out
		}

		if cliConfig.Info {
			info, err := c.Api.Buildinfo(context.Background())
			if err != nil {
				check.ExitError(err)
			}

			output += "\n\n" +
				"Version: " + info.Version + "\n" +
				"Branch: " + info.Branch + "\n" +
				"BuildDate: " + info.BuildDate + "\n" +
				"BuildUser: " + info.BuildUser + "\n" +
				"Revision: " + info.Revision
		}

		check.ExitRaw(rc, output)
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)

	fs := healthCmd.Flags()
	fs.BoolVarP(&cliConfig.PReady, "ready", "r", false, "")
	fs.BoolVarP(&cliConfig.Info, "info", "i", false, "")

	healthCmd.DisableFlagsInUseLine = true
}
