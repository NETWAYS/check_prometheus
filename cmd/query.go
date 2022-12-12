package cmd

import (
	"fmt"
	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	goresult "github.com/NETWAYS/go-check/result"
	"github.com/prometheus/common/model"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

type QueryConfig struct {
	RawQuery string
	Warning  string
	Critical string
	ShowAll  bool
	UnixTime bool
}

type User struct {
	Name       string
	Occupation string
}

var cliQueryConfig QueryConfig

func generateMetricOutput(rc int, metric string, value string) string {
	// Format the metric and RC output for console output
	return fmt.Sprintf(" \\_[%s] %s - value: %s\n", check.StatusText(rc), metric, value)
}

func generatePerfdata(metric string, value string) perfdata.Perfdata {
	// We trim the trailing "} from the string, so that the Perfdata won't have a trailing _
	return perfdata.Perfdata{
		Label: strings.ReplaceAll(strings.TrimRight(metric, "\"}"), " ", ""),
		Value: value,
	}
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Checks the status of a Prometheus query",
	Long: `Checks the status of a Prometheus query and evaluates the result of the alert.
Note: Time range values e.G. 'go_memstats_alloc_bytes_total[0s]' only the latest value will be evaluated, other values will be ignored!`,
	Example: `
	$ check_prometheus query -q 'go_gc_duration_seconds_count' -c 5000 -w 2000
	CRITICAL - 2 Metrics: 1 Critical - 0 Warning - 1 Ok
	 \_[OK] go_gc_duration_seconds_count{instance="localhost:9090", job="prometheus"} - value: 1599
	 \_[CRITICAL] go_gc_duration_seconds_count{instance="node-exporter:9100", job="node-exporter"} - value: 79610
	 | value_go_gc_duration_seconds_count_localhost:9090_prometheus=1599 value_go_gc_duration_seconds_count_node-exporter:9100_node-exporter=79610`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if cliQueryConfig.Warning == "" || cliQueryConfig.Critical == "" {
			check.ExitError(fmt.Errorf("Please specify warning and critical thresholds"))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		var (
			rc             int
			states         []int
			metricOutput   string
			summary        string
			metricsCounter int
			critCounter    int
			warnCounter    int
			okCounter      int
			perfList       perfdata.PerfdataList
		)

		crit, err := check.ParseThreshold(cliQueryConfig.Critical)
		if err != nil {
			check.ExitError(err)
		}

		warn, err := check.ParseThreshold(cliQueryConfig.Warning)
		if err != nil {
			check.ExitError(err)
		}

		c := cliConfig.NewClient()
		err = c.Connect()
		if err != nil {
			check.ExitError(err)
		}

		ctx, cancel := cliConfig.timeoutContext()
		defer cancel()

		result, warnings, err := c.Api.Query(ctx, cliQueryConfig.RawQuery, time.Now())
		if err != nil {
			check.ExitError(err)
		}

		//valType := query.ValType{}
		switch result.Type() {
		// model.ValNone
		default:
			check.ExitError(fmt.Errorf("None value results are not supported"))
		// Scalar - a simple numeric floating point value
		case model.ValScalar:
			check.ExitError(fmt.Errorf("Scalar value results are not supported"))
		case model.ValString:
			// String - a simple string value; currently unused
			check.ExitError(fmt.Errorf("String value results are not supported"))
		case model.ValVector:
			// Instant vector - a set of time series containing a single sample for each time series, all sharing the same timestamp
			vectorVal := result.(model.Vector)

			for _, sample := range vectorVal {
				metricsCounter++

				if crit.DoesViolate(float64(sample.Value)) {
					rc = check.Critical
					critCounter++
				} else if warn.DoesViolate(float64(sample.Value)) {
					rc = check.Warning
					warnCounter++
				} else {
					rc = check.OK
					okCounter++
				}

				states = append(states, rc)
				// Format the metric and RC output for console output
				metricOutput += generateMetricOutput(rc, sample.Metric.String(), sample.Value.String())

				// Generate Perfdata from API return
				perf := generatePerfdata(sample.Metric.String(), sample.Value.String())
				perfList.Add(&perf)
			}
		case model.ValMatrix:
			// Range vector - a set of time series containing a range of data points over time for each time series -> Matrix
			// An example query for a matrix 'go_goroutines{job="prometheus"}[5m]'

			// Note: Only the latest value will be evaluated, other values will be ignored!
			matrixVal := result.(model.Matrix)

			for _, samplestream := range matrixVal {
				metricsCounter++
				samplepair := samplestream.Values[len(samplestream.Values)-1]

				if crit.DoesViolate(float64(samplepair.Value)) {
					rc = check.Critical
					critCounter++
				} else if warn.DoesViolate(float64(samplepair.Value)) {
					rc = check.Warning
					warnCounter++
				} else {
					rc = check.OK
					okCounter++
				}

				states = append(states, rc)
				// Format the metric and RC output for console output
				metricOutput += generateMetricOutput(rc, samplepair.String(), samplepair.Value.String())

				// Generate Perfdata from API return
				perf := generatePerfdata(samplestream.Metric.String(), samplepair.Value.String())
				perfList.Add(&perf)
			}
		}

		// The worst state of all metrics determines the final return state. Example:
		// OK, OK, OK > OK
		// Critical, OK, OK > Critical
		worstState := goresult.WorstState(states...)
		if worstState == check.OK {
			summary += fmt.Sprintf("%d Metrics OK", metricsCounter)
		} else {
			summary += fmt.Sprintf("%d Metrics: %d Critical - %d Warning - %d Ok\n", metricsCounter, critCounter, warnCounter, okCounter)
			summary += metricOutput
		}

		// Should be printed after the Check Plugin output
		// Defer doesn't work because of the os.Exit
		if len(warnings) > 0 {
			summary += fmt.Sprintf("HTTP Warnings: %v\n", strings.Join(warnings, ", "))
		}

		check.ExitRaw(goresult.WorstState(states...), summary, "|", perfList.String())
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)
	fs := queryCmd.Flags()
	fs.StringVarP(&cliQueryConfig.RawQuery, "query", "q", "",
		"An Prometheus query which will be performed and the value result will be evaluated")
	fs.BoolVar(&cliQueryConfig.ShowAll, "show-all", false,
		"Displays all metrics regardless of the status")
	_ = fs.MarkHidden("show-all")
	fs.StringVarP(&cliQueryConfig.Warning, "warning", "w", "10",
		"The warning threshold for a value")
	fs.StringVarP(&cliQueryConfig.Critical, "critical", "c", "20",
		"The critical threshold for a value")

	fs.SortFlags = false
	_ = queryCmd.MarkFlagRequired("query")
}
