package cmd

import (
	"fmt"
	"github.com/NETWAYS/check_prometheus/internal/alert"
	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	"github.com/NETWAYS/go-check/result"
	"github.com/spf13/cobra"
)

func generateOutput(rl alert.Rule, cfg AlertConfig) (output string) {
	// Generates the console output for the AlertingRules
	if len(cfg.AlertName) > 1 || cfg.Group == nil {
		output += " \\_"
	}

	output += fmt.Sprintf("[%s] %s", check.StatusText(rl.GetStatus()), rl.GetOutput())

	if len(cfg.AlertName) > 1 || cfg.Group == nil {
		output += "\n"
	}

	return output
}

func contains(s string, list []string) bool {
	// Tiny helper to see if a string is in a list of strings
	for _, elem := range list {
		if s == elem {
			return true
		}
	}

	return false
}

var alertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Checks the status of a Prometheus alert",
	Long: `Checks the status of a Prometheus alert and evaluates the status of the alert:
firing = 2
pending = 1
inactive = 0`,
	Example: `
	$ check_prometheus alert --name "PrometheusAlertmanagerJobMissing"
	CRITICAL - 1 Alerts: 1 Firing - 0 Pending - 0 Inactive
	 \_[CRITICAL] [PrometheusAlertmanagerJobMissing] - Job: [alertmanager] is firing - value: 1.00
	 | firing=1 pending=0 inactive=0

	$ check_prometheus a alert --name "PrometheusAlertmanagerJobMissing" --name "PrometheusTargetMissing"
	CRITICAL - 2 Alerts: 1 Firing - 0 Pending - 1 Inactive
	 \_[OK] [PrometheusTargetMissing] is inactive
	 \_[CRITICAL] [PrometheusAlertmanagerJobMissing] - Job: [alertmanager] is firing - value: 1.00
	 | total=2 firing=1 pending=0 inactive=1`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			states          []int
			output          string
			summary         string
			counterFiring   int
			counterPending  int
			counterInactive int
			perfList        perfdata.PerfdataList
		)

		c := cliConfig.NewClient()
		err := c.Connect()
		if err != nil {
			check.ExitError(err)
		}

		ctx, cancel := cliConfig.timeoutContext()
		defer cancel()
		// We use the Rules endpoint since it contains
		// the state of inactive Alert Rules, unlike the Alert endpoint
		// Search requested Alert in all Groups and all Rules
		alerts, err := c.Api.Rules(ctx)
		if err != nil {
			check.ExitError(err)
		}

		// Get all rules from all groups into a single list
		rules := alert.FlattenRules(alerts.Groups)

		// Set initial capacity to reduce memory allocations
		var l int
		for _, rl := range rules {
			l = l * len(rl.AlertingRule.Alerts)
		}
		rStates := make([]int, 0, l)

		for _, rl := range rules {

			// If it's not the Alert we're looking for, Skip!
			if cliAlertConfig.AlertName != nil {
				if !contains(rl.AlertingRule.Name, cliAlertConfig.AlertName) {
					continue
				}
			}

			// Skip inactive alerts if flag is set
			if len(rl.AlertingRule.Alerts) == 0 && cliAlertConfig.ProblemsOnly {
				continue
			}

			// Handle Inactive Alerts
			if len(rl.AlertingRule.Alerts) == 0 {
				// Counting states for perfdata
				switch rl.GetStatus() {
				case 0:
					counterInactive++
				case 1:
					counterPending++
				case 2:
					counterFiring++
				}

				// Gather the state to evaluate the worst at the end
				rStates = append(states, rl.GetStatus())
				output += generateOutput(rl, cliAlertConfig)
			}

			// Handle active alerts
			if len(rl.AlertingRule.Alerts) > 0 {
				// Handle Pending or Firing Alerts
				for _, alert := range rl.AlertingRule.Alerts {
					// Counting states for perfdata
					switch rl.GetStatus() {
					case 0:
						counterInactive++
					case 1:
						counterPending++
					case 2:
						counterFiring++
					}

					rl.Alert = alert
					// Gather the state to evaluate the worst at the end
					rStates = append(states, rl.GetStatus())
					output += generateOutput(rl, cliAlertConfig)
				}
			}
		}
		states = rStates

		counterAlert := counterFiring + counterPending + counterInactive
		if len(cliAlertConfig.AlertName) > 1 || counterAlert > 1 {
			perfList = perfdata.PerfdataList{
				{Label: "total", Value: counterAlert},
				{Label: "firing", Value: counterFiring},
				{Label: "pending", Value: counterPending},
				{Label: "inactive", Value: counterInactive},
			}
		}

		if len(cliAlertConfig.AlertName) == 1 && counterAlert == 1 {
			perfList = perfdata.PerfdataList{
				{Label: "firing", Value: counterFiring},
				{Label: "pending", Value: counterPending},
				{Label: "inactive", Value: counterInactive},
			}
		}

		summary += fmt.Sprintf("%d Alerts: %d Firing - %d Pending - %d Inactive",
			counterAlert,
			counterFiring,
			counterPending,
			counterInactive)

		if result.WorstState(states...) == 0 {
			check.ExitRaw(result.WorstState(states...), "Alerts inactive", "|", perfList.String())
		} else {
			check.ExitRaw(result.WorstState(states...), summary+"\n"+output, "|", perfList.String())
		}
	},
}

func init() {
	rootCmd.AddCommand(alertCmd)
	fs := alertCmd.Flags()
	fs.StringSliceVarP(&cliAlertConfig.AlertName, "name", "n", nil,
		"The name of one or more specific alerts to check."+
			"\nThis parameter can be repeated e.G.: '--name alert1 --name alert2'"+
			"\nIf no name is given, all alerts will be evaluated")
	fs.BoolVarP(&cliAlertConfig.ProblemsOnly, "problems", "P", false,
		"Display only alerts which status is not inactive/OK. Note that in combination with the --name flag this might result in no Alerts being displayed")
}
