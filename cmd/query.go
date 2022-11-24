package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	goresult "github.com/NETWAYS/go-check/result"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/spf13/cobra"
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

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Checks the status of a Prometheus query",
	Long: `Checks the status of a Prometheus query and evaluates the result of the alert.
Note: Time range values e.G. 'go_memstats_alloc_bytes_total[0s]' only the latest value will be evaluated, other values will be ignored!`,
	Example: `  $ check_prometheus query -q 'go_gc_duration_seconds_count' -c 5000 -w 2000   
  CRITICAL - 2 Metrics: 1 Critical - 0 Warning - 1 Ok
   \_[OK] go_gc_duration_seconds_count{instance="localhost:9090", job="prometheus"} - value: 1599
   \_[CRITICAL] go_gc_duration_seconds_count{instance="node-exporter:9100", job="node-exporter"} - value: 79610
   | value_go_gc_duration_seconds_count_localhost:9090_prometheus=1599 value_go_gc_duration_seconds_count_node-exporter:9100_node-exporter=79610`,
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
		switch result.Type() {
		// Scalar - a simple numeric floating point value
		case model.ValScalar:
			check.ExitError(fmt.Errorf("scalar value results are not supported"))
		case model.ValVector:
			// Instant vector - a set of time series containing a single sample for each time series, all sharing the same timestamp
			vectorVal := result.(model.Vector)

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

				states = append(states, rc)

				output = fmt.Sprintf(" \\_[%s] %s - value: %s\n", check.StatusText(rc), metric.Metric.String(), metric.Value)

				metricOutput += output
				output = ""

				// Marshalls the Prometheus output -> go_gc_duration_seconds_count{instance="node-exporter:9100", job="node-exporter"}
				// Need the JSON to fill the performance data with the labels
				jsn, err := metric.MarshalJSON()
				if err != nil {
					check.ExitError(err)
				}

				var res map[string]interface{}

				err = json.Unmarshal(jsn, &res)
				if err != nil {
					check.ExitError(err)
				}

				// map[metric:map[__name__:go_goroutines instance:localhost:9090 job:prometheus] value:[1.669290557403e+09 38]]
				mtr := res["metric"].(map[string]interface{})
				var (
					metrName string
					instance string
					job      string
				)
				for key, val := range mtr {
					switch key {
					case "__name__":
						metrName = fmt.Sprintf("%v", val)
					case "instance":
						instance = fmt.Sprintf("%v", val)
					case "job":
						job = fmt.Sprintf("%v", val)
					}
				}

				perf := perfdata.Perfdata{
					Label: fmt.Sprintf("value_%s_%s_%s", metrName, instance, job),
					Value: metric.Value,
				}

				perfList.Add(&perf)
			}
		case model.ValMatrix:
			// Range vector - a set of time series containing a range of data points over time for each time series -> Matrix
			// An example query for a matrix 'go_goroutines{job="prometheus"}[5m]'

			// Note: Only the latest value will be evaluated, other values will be ignored!
			matrixVal := result.(model.Matrix)

			for _, val := range matrixVal {
				metricsCounter++
				v := val.Values[len(val.Values)-1]

				if crit.DoesViolate(float64(v.Value)) {
					rc = check.Critical
					critCounter++
				} else if warn.DoesViolate(float64(v.Value)) {
					rc = check.Warning
					warnCounter++
				} else {
					rc = check.OK
					okCounter++
				}

				states = append(states, rc)

				output = fmt.Sprintf(" \\_[%s] %s - value: %s\n", check.StatusText(rc), val.Metric.String(), v.Value)

				metricOutput += output
				output = ""

				var (
					metrName string
					instance string
					job      string
				)
				for key, value := range val.Metric {
					switch key {
					case "__name__":
						metrName = fmt.Sprintf("%v", value)
					case "instance":
						instance = fmt.Sprintf("%v", value)
					case "job":
						job = fmt.Sprintf("%v", value)
					}
				}

				perf := perfdata.Perfdata{
					Label: fmt.Sprintf("value_%s_%s_%s", metrName, instance, job),
					Value: v.Value,
				}

				perfList.Add(&perf)
			}
		case model.ValString:
			// String - a simple string value; currently unused
			check.ExitError(fmt.Errorf("string value results are not supported"))
		default:
			// model.ValNone
			check.ExitError(fmt.Errorf("none value results are not supported"))
		}

		worstState := goresult.WorstState(states...)
		if worstState == check.OK {
			summary += fmt.Sprintf("%d Metrics OK", metricsCounter)
		} else {
			summary += fmt.Sprintf("%d Metrics: %d Critical - %d Warning - %d Ok\n", metricsCounter, critCounter, warnCounter, okCounter)
			summary += metricOutput

		}

		// Should be printed after the Check Plugin output
		defer printWarning(warnings)

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

func printWarning(warnings v1.Warnings) {
	// TODO Validate this
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
}
