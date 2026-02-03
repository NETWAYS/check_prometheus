package cmd

import (
	"fmt"
	"slices"

	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	"github.com/NETWAYS/go-check/result"
	"github.com/spf13/cobra"
)

type AlertmanagerConfig struct {
	AlertName     []string
	ExcludeAlerts []string
	ExcludeLabels []string
	IncludeLabels []string
	ProblemsOnly  bool
	NoAlertsState string
}

const stateUnprocessed = "unprocessed"
const stateActive = "active"
const stateSuppressed = "suppressed"

var cliAlertmanagerConfig AlertmanagerConfig

var alertmanagerCmd = &cobra.Command{
	Use:   "alertmanager",
	Short: "Checks the status of a Alertmanager alert",
	Long:  `Checks the status of a Alertmanager alert and evaluates the status of the alert`,
	Run: func(_ *cobra.Command, _ []string) {
		// Convert --no-alerts-state to integer and validate input
		noAlertsState, err := convertStateToInt(cliAlertmanagerConfig.NoAlertsState)
		if err != nil {
			check.ExitError(fmt.Errorf("invalid value for --no-alerts-state: %s", cliAlertmanagerConfig.NoAlertsState))
		}

		var (
			counterUnprocessed int
			counterActive      int
			counterSuppressed  int
		)

		c := cliConfig.NewClient()
		errCon := c.Connect()

		if errCon != nil {
			check.ExitError(errCon)
		}

		ctx, cancel := cliConfig.timeoutContext()
		defer cancel()

		alerts, err := c.GetAlertmanagerAlerts(ctx)
		if err != nil {
			check.ExitError(err)
		}

		// If there are no alerts we can exit early
		if len(alerts) == 0 {
			// Just an empty PerfdataList to have consistent perfdata output
			pdlist := perfdata.PerfdataList{
				{Label: "total", Value: 0},
				{Label: "unprocessed", Value: 0},
				{Label: "active", Value: 0},
				{Label: "suppressed", Value: 0},
			}

			// Since the user is expecting the state of a certain alert and
			// it that is not present it might be noteworthy.
			if cliAlertmanagerConfig.AlertName != nil {
				check.ExitRaw(check.Unknown, "No such alert defined", "|", pdlist.String())
			}

			check.ExitRaw(noAlertsState, "No alerts found", "|", pdlist.String())
		}

		var overall result.Overall

		for _, al := range alerts {
			// If it's not the Alert we're looking for, Skip!
			if cliAlertmanagerConfig.AlertName != nil {
				if !slices.Contains(cliAlertmanagerConfig.AlertName, al.GetName()) {
					continue
				}
			}

			labelsMatchedInclude := matchesLabel(al.Labels, cliAlertmanagerConfig.IncludeLabels)

			if len(cliAlertConfig.IncludeLabels) > 0 && !labelsMatchedInclude {
				// If the alert labels don't match here we can skip it.
				continue
			}

			alertMatchedExclude, regexErr := matches(al.GetName(), cliAlertmanagerConfig.ExcludeAlerts)

			if regexErr != nil {
				check.ExitRaw(check.Unknown, "Invalid regular expression provided:", regexErr.Error())
			}

			if alertMatchedExclude {
				// If the alert matches a regex from the list we can skip it.
				continue
			}

			labelsMatchedExclude := matchesLabel(al.Labels, cliAlertConfig.ExcludeLabels)

			if len(cliAlertConfig.ExcludeLabels) > 0 && labelsMatchedExclude {
				// If the alert labels matches here we can skip it.
				continue
			}

			sc := result.NewPartialResult()

			switch al.GetState() {
			case stateUnprocessed:
				//nolint: errcheck
				sc.SetState(check.Warning)
				sc.Output = al.GetName() + " is unprocessed"
				counterUnprocessed++
			case stateActive:
				//nolint: errcheck
				sc.SetState(check.Critical)
				sc.Output = al.GetName() + " is active"
				counterActive++
			case stateSuppressed:
				//nolint: errcheck
				sc.SetState(check.OK)
				sc.Output = al.GetName() + " is suppressed"
				counterSuppressed++
			default:
				//nolint: errcheck
				sc.SetState(check.Unknown)
				sc.Output = al.GetName() + "invalid alert state"
			}

			overall.AddSubcheck(sc)
		}

		counterAlert := counterUnprocessed + counterActive + counterSuppressed

		perfList := perfdata.PerfdataList{
			{Label: "total", Value: counterAlert},
			{Label: "unprocessed", Value: counterUnprocessed},
			{Label: "active", Value: counterActive},
			{Label: "suppressed", Value: counterSuppressed},
		}

		// When there are no alerts we add an empty PartialResult just to have consistent output
		if len(overall.PartialResults) == 0 {
			sc := result.NewPartialResult()
			// We already make sure it's valid
			//nolint: errcheck
			sc.SetDefaultState(noAlertsState)
			sc.Output = "No alerts retrieved"
			overall.AddSubcheck(sc)
		}

		overall.PartialResults[0].Perfdata = append(overall.PartialResults[0].Perfdata, perfList...)

		overall.Summary = fmt.Sprintf("%d total alerts: %d unprocessed - %d active - %d suppressed",
			counterAlert,
			counterUnprocessed,
			counterActive,
			counterSuppressed)

		check.ExitRaw(overall.GetStatus(), overall.GetOutput())
	},
}

func init() {
	rootCmd.AddCommand(alertmanagerCmd)

	fs := alertmanagerCmd.Flags()

	fs.StringVarP(&cliAlertmanagerConfig.NoAlertsState, "no-alerts-state", "T", "OK", "State to assign when no alerts are found (0, 1, 2, 3, OK, WARNING, CRITICAL, UNKNOWN). If not set this defaults to OK")

	fs.StringArrayVar(&cliAlertmanagerConfig.ExcludeAlerts, "exclude-alert", []string{}, "Alerts to ignore. Can be used multiple times and supports regex.")

	fs.StringSliceVarP(&cliAlertmanagerConfig.AlertName, "name", "n", nil,
		"The name of one or more specific alerts to check."+
			"\nThis parameter can be repeated e.G.: '--name alert1 --name alert2'"+
			"\nIf no name is given, all alerts will be evaluated")

	fs.StringArrayVar(&cliAlertmanagerConfig.IncludeLabels, "include-label", []string{},
		"The label of one or more specific alerts to include. "+
			"\nThis parameter can be repeated e.g.: '--include-label prio=high --include-label another=example'"+
			"\nNote that repeated --include-label are combined using a union.")

	fs.StringArrayVar(&cliAlertmanagerConfig.ExcludeLabels, "exclude-label", []string{},
		"The label of one or more specific alerts to exclude."+
			"\nThis parameter can be repeated e.g.: '--exclude-label prio=high --exclude-label another=example'")
}
