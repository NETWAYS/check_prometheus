package cmd

import (
	"github.com/NETWAYS/go-check"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/spf13/cobra"
)

const (
	// Possible values for Alert States.
	StateFiring   = "firing"
	StateInactive = "inactive"
	StatePending  = "pending"
)

type AlertConfig struct {
	Name string
}

var cliAlertConfig AlertConfig

var alertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Checks the status of an Prometheus alert",
	Long: `Checks the status of an Prometheus alert

	1. --name

The alert status is:
	inactive = OK
	pending = WARNING
	firing = CRITICAL`,
	Run: func(cmd *cobra.Command, args []string) {
		c := cliConfig.Client()
		err := c.Connect()

		if err != nil {
			check.ExitError(err)
		}

		ctx, cancel := cliConfig.timeoutContext()
		defer cancel()
		result, err := c.Api.Rules(ctx)

		if err != nil {
			check.ExitError(err)
		}

		// Search requested Alert in all Groups and all Rules
		for _, grp := range result.Groups {
			for _, rl := range grp.Rules {
				// TODO Only do this on AlertingRule not on RecordingRule
				rule := rl.(v1.AlertingRule)
				if rule.Name == cliAlertConfig.Name {
					switch rule.State {
					case StateInactive:
						check.Exitf(check.OK, "Alert %s inactive", cliAlertConfig.Name)
					case StatePending:
						check.Exitf(check.Warning, "Alert %s pending", cliAlertConfig.Name)
					case StateFiring:
						check.Exitf(check.Critical, "Alert %s firing", cliAlertConfig.Name)
					}
				}
			}
		}
		check.Exitf(check.Unknown, "Alert %s not found", cliAlertConfig.Name)
	},
}

func init() {
	rootCmd.AddCommand(alertCmd)
	fs := alertCmd.Flags()
	fs.StringVarP(&cliAlertConfig.Name, "name", "n", "",
		"The name of the alert to check")

	_ = alertCmd.MarkFlagRequired("name")
}
