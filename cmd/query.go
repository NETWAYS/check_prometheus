package cmd

import (
	"fmt"
	"github.com/NETWAYS/go-check"
	goresult "github.com/NETWAYS/go-check/result"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/spf13/cobra"
	"time"
)

type QueryConfig struct {
	RawQuery string
	NaNOK    bool
	EmptyOK  bool
	Warning  string
	Critical string
	ShowAll  bool
	UnixTime bool
}

var cliQueryConfig QueryConfig

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Checks the status of an Prometheus query",
	Long: `Checks the status of an Prometheus query

	1. --query`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if cliQueryConfig.Warning == "" || cliQueryConfig.Critical == "" {
			check.ExitError(fmt.Errorf("please specify warning and critical thresholds"))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		var (
			rc             int
			states         []int
			output         string
			metricOutput   string
			summary        string
			metricsCounter int
			critCounter    int
			warnCounter    int
			okCounter      int
		)

		crit, err := check.ParseThreshold(cliQueryConfig.Critical)
		if err != nil {
			check.ExitError(err)
		}

		warn, err := check.ParseThreshold(cliQueryConfig.Warning)
		if err != nil {
			check.ExitError(err)
		}

		c := cliConfig.Client()
		err = c.Connect()
		if err != nil {
			check.ExitError(err)
		}

		ctx, cancel := cliConfig.timeoutContext()
		defer cancel()

		// TODO Should time be a flag?

		result, warnings, err := c.Api.Query(ctx, cliQueryConfig.RawQuery, time.Now())
		if err != nil {
			check.ExitError(err)
		}

		//valType := query.ValType{}
		summary += fmt.Sprintf("Found ")

		switch result.Type() {
		// Scalar - a simple numeric floating point value
		case model.ValScalar:
			//fmt.Println("SCALAR")
			//scalarVal := result.(*model.Scalar)
			//fmt.Println(scalarVal)
		case model.ValVector:
			// Instant vector - a set of time series containing a single sample for each time series, all sharing the same timestamp
			// An example query for SINGLE Vector:  'go_goroutines{job="prometheus"}'
			// An example query for MULTIPLE Vector: 'go_goroutines'
			//fmt.Println("VECTOR")
			vectorVal := result.(model.Vector)
			//valType.Vector = result.(model.Vector)

			for _, metric := range vectorVal {
				metricsCounter++
				if crit.DoesViolate(float64(metric.Value)) {
					rc = check.Critical
					critCounter++
				} else if warn.DoesViolate(float64(metric.Value)) {
					rc = check.Warning
					warnCounter++
				} else {
					rc = check.OK
					okCounter++
				}

				if len(vectorVal) > 1 {
					metricOutput += fmt.Sprintf("[%s] ", check.StatusText(rc))
				}

				metricOutput += fmt.Sprintf("%s\n", metric.Metric.String())
				states = append(states, rc)

				output += fmt.Sprintf(" \\_ %s @ %s\n", metric.Value, metric.Timestamp.Time())
				metricOutput += output
			}
		case model.ValMatrix:
			// Range vector - a set of time series containing a range of data points over time for each time series -> Matrix
			// An example query for a matrix 'go_goroutines{job="prometheus"}[5m]'
			//fmt.Println("MATRIX")
			matrixVal := result.(model.Matrix)

			for _, val := range matrixVal {
				metricsCounter++
				for _, metric := range val.Values {
					if crit.DoesViolate(float64(metric.Value)) {
						rc = check.Critical
					} else if warn.DoesViolate(float64(metric.Value)) {
						rc = check.Warning
					} else {
						rc = check.OK
					}

					states = append(states, rc)

					output += fmt.Sprintf(" \\_ %s @ %s\n", metric.Value, metric.Timestamp.Time())
				}

				worstState := goresult.WorstState(states...)
				metricOutput += fmt.Sprintf("[%s] %s\n", check.StatusText(worstState), val.Metric.String())
				metricOutput += output

				switch worstState {
				case check.Critical:
					critCounter++
				case check.Warning:
					warnCounter++
				case check.OK:
					okCounter++
				}
			}
		case model.ValString:
			// String - a simple string value; currently unused
		default:
			// model.ValNone
			//fmt.Println("NONE")
		}

		worstState := goresult.WorstState(states...)
		if worstState == check.OK {
			summary += fmt.Sprintf("%d Metrics - all Metrics Ok", metricsCounter)
		} else {
			if metricsCounter == 1 {
				summary = metricOutput
			} else {
				summary += fmt.Sprintf("%d Metrics - %d Critical - %d Warning - %d Ok\n", metricsCounter, critCounter, warnCounter, okCounter)
				summary += metricOutput
			}
		}

		// Should be printed after the Check Plugin output
		defer printWarning(warnings)

		check.ExitRaw(goresult.WorstState(states...), summary)
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)
	fs := queryCmd.Flags()
	fs.StringVarP(&cliQueryConfig.RawQuery, "query", "n", "",
		"The query to check")
	fs.BoolVar(&cliQueryConfig.NaNOK, "nan-ok", false,
		"Accept NaN as an \"OK\" result")
	fs.BoolVar(&cliQueryConfig.EmptyOK, "empty-ok", false,
		"Accept an empty vector (null) as an \"OK\" result")
	fs.BoolVar(&cliQueryConfig.ShowAll, "show-all", false, "Displays all metrics regardless of the status")
	fs.StringVarP(&cliQueryConfig.Critical, "critical", "c", "", "")
	fs.StringVarP(&cliQueryConfig.Warning, "warning", "w", "", "")

	_ = queryCmd.MarkFlagRequired("query")
}

func printWarning(warnings v1.Warnings) {
	// TODO Validate this
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
}
