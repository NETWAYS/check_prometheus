package cmd

import (
	"fmt"
	"github.com/NETWAYS/check_prometheus/internal/alert"
	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	"github.com/NETWAYS/go-check/result"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/spf13/cobra"
)

var alertCmd = &cobra.Command{
	Use:   "alert",
	Short: "Checks the status of a Prometheus alert",
	Long: `Checks the status of a Prometheus alert and evaluates the status of the alert:
firing = 2
pending = 1
inactive = 0`,
	Example: `  $ check_prometheus alert --name "PrometheusAlertmanagerJobMissing"
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
			counterAlert    int
			counterFiring   int
			counterPending  int
			counterInactive int
			perfList        perfdata.PerfdataList
			//value           int
		)

		rule := alert.Alertingrule{}

		c := cliConfig.Client()
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

		// TODO SUPER SUPER hacky, needs to be refactored ASAP!
		// Search for specified groups
		for _, grp := range alerts.Groups {
			for _, rl := range grp.Rules {

				if _, ok := rl.(v1.AlertingRule); ok {
					rule.AlertingRule = rl.(v1.AlertingRule)
				}

				// Search for specified alert names
				if cliAlertConfig.AlertName == nil {
					if len(rule.AlertingRule.Alerts) == 0 {
						counterAlert++

						switch rule.GetStatus() {
						case 0:
							counterInactive++
						case 1:
							counterPending++
						case 2:
							counterFiring++
						}

						states = append(states, rule.GetStatus())

						if len(cliAlertConfig.AlertName) > 1 || cliAlertConfig.Group == nil {
							output += " \\_"
						}

						output += fmt.Sprintf("[%s] %s", check.StatusText(rule.GetStatus()), rule.GetOutput())

						if len(cliAlertConfig.AlertName) > 1 || cliAlertConfig.Group == nil {
							output += "\n"
						}
					} else {
						for _, alert := range rule.AlertingRule.Alerts {
							counterAlert++

							switch rule.GetStatus() {
							case 0:
								counterInactive++
							case 1:
								counterPending++
							case 2:
								counterFiring++
							}

							rule.Alert = alert
							states = append(states, rule.GetStatus())

							if len(cliAlertConfig.AlertName) > 1 || cliAlertConfig.Group == nil {
								output += " \\_"
							}

							output += fmt.Sprintf("[%s] %s", check.StatusText(rule.GetStatus()), rule.GetOutput())

							if len(cliAlertConfig.AlertName) > 1 || cliAlertConfig.Group == nil {
								output += "\n"
							}
						}
					}
				} else {
					for _, name := range cliAlertConfig.AlertName {
						if name == rule.AlertingRule.Name {
							if len(rule.AlertingRule.Alerts) == 0 {
								counterAlert++

								switch rule.GetStatus() {
								case 0:
									counterInactive++
								case 1:
									counterPending++
								case 2:
									counterFiring++
								}

								states = append(states, rule.GetStatus())

								if len(cliAlertConfig.AlertName) > 1 || cliAlertConfig.Group == nil {
									output += " \\_"
								}

								output += fmt.Sprintf("[%s] %s", check.StatusText(rule.GetStatus()), rule.GetOutput())

								if len(cliAlertConfig.AlertName) > 1 || cliAlertConfig.Group == nil {
									output += "\n"
								}
							} else {
								for _, alert := range rule.AlertingRule.Alerts {
									counterAlert++

									switch rule.GetStatus() {
									case 0:
										counterInactive++
									case 1:
										counterPending++
									case 2:
										counterFiring++
									}

									rule.Alert = alert
									states = append(states, rule.GetStatus())

									if len(cliAlertConfig.AlertName) > 1 || cliAlertConfig.Group == nil {
										output += " \\_"
									}

									output += fmt.Sprintf("[%s] %s", check.StatusText(rule.GetStatus()), rule.GetOutput())

									if len(cliAlertConfig.AlertName) > 1 || cliAlertConfig.Group == nil {
										output += "\n"
									}
								}
							}
						}
					}
				}
			}
		}

		if len(cliAlertConfig.AlertName) > 1 || counterAlert > 1 {
			perfList = perfdata.PerfdataList{
				{Label: "total", Value: counterAlert},
				{Label: "firing", Value: counterFiring},
				{Label: "pending", Value: counterPending},
				{Label: "inactive", Value: counterInactive},
			}
		} else if len(cliAlertConfig.AlertName) == 1 && counterAlert == 1 {
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
	fs.BoolVarP(&cliAlertConfig.Problems, "problems", "P", false,
		"Displays only alerts which status is not OK")
	// TODO: has to be implemented
	_ = fs.MarkHidden("problems")
}
