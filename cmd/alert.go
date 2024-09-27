package cmd

import (
	"fmt"

	"github.com/NETWAYS/check_prometheus/internal/alert"
	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	"github.com/NETWAYS/go-check/result"
	"github.com/spf13/cobra"
)

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

	$ check_prometheus alert --name "PrometheusAlertmanagerJobMissing" --name "PrometheusTargetMissing"
	CRITICAL - 2 Alerts: 1 Firing - 0 Pending - 1 Inactive
	 \_[OK] [PrometheusTargetMissing] is inactive
	 \_[CRITICAL] [PrometheusAlertmanagerJobMissing] - Job: [alertmanager] is firing - value: 1.00
	 | total=2 firing=1 pending=0 inactive=1`,
	Run: func(_ *cobra.Command, _ []string) {
		var (
			counterFiring   int
			counterPending  int
			counterInactive int
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
		alerts, err := c.API.Rules(ctx)
		if err != nil {
			check.ExitError(err)
		}

		// Get all rules from all groups into a single list
		rules := alert.FlattenRules(alerts.Groups)

		// Set initial capacity to reduce memory allocations
		var l int
		for _, rl := range rules {
			l *= len(rl.AlertingRule.Alerts)
		}

		var overall result.Overall

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

				sc := result.NewPartialResult()

				_ = sc.SetState(rl.GetStatus())
				sc.Output = rl.GetOutput()
				overall.AddSubcheck(sc)
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

					sc := result.NewPartialResult()

					_ = sc.SetState(rl.GetStatus())
					// Set the alert in the internal Type to generate the output
					rl.Alert = alert
					sc.Output = rl.GetOutput()
					overall.AddSubcheck(sc)
				}
			}
		}

		counterAlert := counterFiring + counterPending + counterInactive
		if len(cliAlertConfig.AlertName) > 1 || counterAlert > 1 {
			perfList := perfdata.PerfdataList{
				{Label: "total", Value: counterAlert},
				{Label: "firing", Value: counterFiring},
				{Label: "pending", Value: counterPending},
				{Label: "inactive", Value: counterInactive},
			}

			overall.PartialResults[0].Perfdata = append(overall.PartialResults[0].Perfdata, perfList...)
		}

		if len(cliAlertConfig.AlertName) == 1 && counterAlert == 1 {
			perfList := perfdata.PerfdataList{
				{Label: "firing", Value: counterFiring},
				{Label: "pending", Value: counterPending},
				{Label: "inactive", Value: counterInactive},
			}
			overall.PartialResults[0].Perfdata = append(overall.PartialResults[0].Perfdata, perfList...)
		}

		overall.Summary = fmt.Sprintf("%d Alerts: %d Firing - %d Pending - %d Inactive",
			counterAlert,
			counterFiring,
			counterPending,
			counterInactive)

		check.ExitRaw(overall.GetStatus(), overall.GetOutput())
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
		"Display only alerts which status is not inactive/OK. Note that in combination with the --name flag this might result in no alerts being displayed")
}
