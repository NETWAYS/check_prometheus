package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/NETWAYS/check_prometheus/internal/alert"
	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	"github.com/NETWAYS/go-check/result"
	"github.com/spf13/cobra"
)

type AlertConfig struct {
	AlertName     []string
	Group         []string
	ExcludeAlerts []string
	ProblemsOnly  bool
	NoAlertsState string
}

var cliAlertConfig AlertConfig

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
		// Convert --no-alerts-state to integer and validate input
		noAlertsState, err := convertStateToInt(cliAlertConfig.NoAlertsState)
		if err != nil {
			check.ExitError(fmt.Errorf("invalid value for --no-alerts-state: %s", cliAlertConfig.NoAlertsState))
		}

		var (
			counterFiring   int
			counterPending  int
			counterInactive int
		)

		c := cliConfig.NewClient()
		err = c.Connect()

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
		rules := alert.FlattenRules(alerts.Groups, cliAlertConfig.Group)

		// If there are no rules we can exit early
		if len(rules) == 0 {
			// Just an empty PerfdataList to have consistent perfdata output
			pdlist := perfdata.PerfdataList{
				{Label: "total", Value: 0},
				{Label: "firing", Value: 0},
				{Label: "pending", Value: 0},
				{Label: "inactive", Value: 0},
			}

			// Since the user is expecting the state of a certain alert and
			// it that is not present it might be noteworthy.
			if cliAlertConfig.AlertName != nil {
				check.ExitRaw(check.Unknown, "No such alert defined", "|", pdlist.String())
			}
			check.ExitRaw(noAlertsState, "No alerts defined", "|", pdlist.String())
		}

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

			alertMatched, regexErr := matches(rl.AlertingRule.Name, cliAlertConfig.ExcludeAlerts)

			if regexErr != nil {
				check.ExitRaw(check.Unknown, "Invalid regular expression provided:", regexErr.Error())
			}

			if alertMatched {
				// If the alert matches a regex from the list we can skip it.
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

		perfList := perfdata.PerfdataList{
			{Label: "total", Value: counterAlert},
			{Label: "firing", Value: counterFiring},
			{Label: "pending", Value: counterPending},
			{Label: "inactive", Value: counterInactive},
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

	fs.StringVarP(&cliAlertConfig.NoAlertsState, "no-alerts-state", "T", "OK", "State to assign when no alerts are found (0, 1, 2, 3, OK, WARNING, CRITICAL, UNKNOWN). If not set this defaults to OK")

	fs.StringArrayVar(&cliAlertConfig.ExcludeAlerts, "exclude-alert", []string{}, "Alerts to ignore. Can be used multiple times and supports regex.")

	fs.StringSliceVarP(&cliAlertConfig.AlertName, "name", "n", nil,
		"The name of one or more specific alerts to check."+
			"\nThis parameter can be repeated e.G.: '--name alert1 --name alert2'"+
			"\nIf no name is given, all alerts will be evaluated")

	fs.StringSliceVarP(&cliAlertConfig.Group, "group", "g", nil,
		"The name of one or more specific groups to check for alerts."+
			"\nThis parameter can be repeated e.G.: '--group group1 --group group2'"+
			"\nIf no group is given, all groups will be scanned for alerts")

	fs.BoolVarP(&cliAlertConfig.ProblemsOnly, "problems", "P", false,
		"Display only alerts which status is not inactive/OK. Note that in combination with the --name flag this might result in no alerts being displayed")
}

// Function to convert state to integer.
func convertStateToInt(state string) (int, error) {
	state = strings.ToUpper(state)
	switch state {
	case "OK", "0":
		return check.OK, nil
	case "WARNING", "1":
		return check.Warning, nil
	case "CRITICAL", "2":
		return check.Critical, nil
	case "UNKNOWN", "3":
		return check.Unknown, nil
	default:
		return check.Unknown, errors.New("invalid state")
	}
}

// Matches a list of regular expressions against a string.
func matches(input string, regexToExclude []string) (bool, error) {
	for _, regex := range regexToExclude {
		re, err := regexp.Compile(regex)

		if err != nil {
			return false, err
		}

		if re.MatchString(input) {
			return true, nil
		}
	}

	return false, nil
}
