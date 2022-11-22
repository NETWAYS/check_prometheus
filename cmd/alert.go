package cmd

import (
	"fmt"
	"github.com/NETWAYS/check_prometheus/internal/alert"
	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/result"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/spf13/cobra"
)

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
		var (
			states          []int
			output          string
			summary         string
			counterAlert    int
			counterFiring   int
			counterPending  int
			counterInactive int
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

		summary += fmt.Sprintf("Found ")

		// TODO SUPER SUPER hacky, needs to be refactored ASAP!
		// Search for specified groups
		for _, grp := range alerts.Groups {
			for _, rl := range grp.Rules {
				rule.Alertingrule = rl.(v1.AlertingRule)
				// Search for specified alert names
				if cliAlertConfig.AlertName == nil {
					if len(rule.Alertingrule.Alerts) == 0 {
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
							output += " \n"
						}
					} else {
						for _, alert := range rule.Alertingrule.Alerts {
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
								output += " \n"
							}
						}
					}
				} else {
					for _, name := range cliAlertConfig.AlertName {
						if name == rule.Alertingrule.Name {
							if len(rule.Alertingrule.Alerts) == 0 {
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
									output += " \n"
								}
							} else {
								for _, alert := range rule.Alertingrule.Alerts {
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
										output += " \n"
									}
								}
							}
						}
					}
				}
			}
		}

		summary += fmt.Sprintf("%d alerts - firing %d - pending %d - inactive %d",
			counterAlert,
			counterFiring,
			counterPending,
			counterInactive)

		if result.WorstState(states...) == 0 {
			check.ExitRaw(result.WorstState(states...), "All alerts are inactive")
		} else {
			check.ExitRaw(result.WorstState(states...), summary+"\n"+output)
		}
	},
}

func init() {
	rootCmd.AddCommand(alertCmd)
	fs := alertCmd.Flags()
	fs.BoolVarP(&cliAlertConfig.Problems, "problems", "P", false,
		"Displays only alerts which status is not OK.")
}
