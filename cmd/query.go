package cmd

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	goresult "github.com/NETWAYS/go-check/result"
	"github.com/prometheus/common/model"
	"github.com/spf13/cobra"
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

var replacer = strings.NewReplacer("{", "_", "}", "", "\"", "", ",", "_", " ", "")

func generateMetricOutput(metric string, value string) string {
	// Format the metric and RC output for console output
	return fmt.Sprintf(" %s - value: %s", metric, value)
}

type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

func generatePerfdata[T Number](metric string, value T, warning, critical *check.Threshold) perfdata.Perfdata {
	// We trim the trailing "} from the string, so that the Perfdata won't have a trailing _
	return perfdata.Perfdata{
		Label: replacer.Replace(metric),
		Value: value,
		Warn:  warning,
		Crit:  critical,
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
			check.ExitError(fmt.Errorf("please specify warning and critical thresholds"))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
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

		result, warnings, err := c.API.Query(ctx, cliQueryConfig.RawQuery, time.Now())

		if err != nil {
			if strings.Contains(err.Error(), "unmarshalerDecoder: unexpected value type \"string\"") {
				err = fmt.Errorf("string value results are not supported")
			}
			check.ExitError(err)
		}

		overall := goresult.Overall{}

		switch result.Type() {
		default:
			check.ExitError(fmt.Errorf("none value results are not supported"))
		// Scalar - a simple numeric floating point value
		case model.ValScalar:
			check.ExitError(fmt.Errorf("scalar value results are not supported"))
		case model.ValNone:
			check.ExitError(fmt.Errorf("none value results are not supported"))
		case model.ValString:
			// String - a simple string value; currently unused
			check.ExitError(fmt.Errorf("string value results are not supported"))
		case model.ValVector:
			// Instant vector - a set of time series containing a single sample for each time series, all sharing the same timestamp
			vectorVal := result.(model.Vector)

			// Set initial capacity to reduce memory allocations
			for _, sample := range vectorVal {

				numberValue := float64(sample.Value)
				partial := goresult.NewPartialResult()

				if crit.DoesViolate(numberValue) {
					partial.SetState(check.Critical)
				} else if warn.DoesViolate(numberValue) {
					partial.SetState(check.Warning)
				} else {
					partial.SetState(check.OK)
				}

				// Format the metric and RC output for console output
				partial.Output = generateMetricOutput(sample.Metric.String(), sample.Value.String())

				// Generate Perfdata from API return
				if math.IsInf(numberValue, 0) || math.IsNaN(numberValue) {
					continue
				}

				perf := generatePerfdata(sample.Metric.String(), numberValue, warn, crit)
				partial.Perfdata.Add(&perf)
				overall.AddSubcheck(partial)
			}

		case model.ValMatrix:
			// Range vector - a set of time series containing a range of data points over time for each time series -> Matrix
			// An example query for a matrix 'go_goroutines{job="prometheus"}[5m]'

			// Note: Only the latest value will be evaluated, other values will be ignored!
			matrixVal := result.(model.Matrix)

			for _, samplestream := range matrixVal {
				samplepair := samplestream.Values[len(samplestream.Values)-1]

				numberValue := float64(samplepair.Value)

				partial := goresult.NewPartialResult()

				if crit.DoesViolate(numberValue) {
					partial.SetState(check.Critical)
				} else if warn.DoesViolate(numberValue) {
					partial.SetState(check.Warning)
				} else {
					partial.SetState(check.OK)
				}

				// Format the metric and RC output for console output
				partial.Output = generateMetricOutput(samplepair.String(), samplepair.Value.String())

				valueString := samplepair.Value.String()

				valueNumber, err := strconv.ParseFloat(valueString, 64)
				if err == nil {
					pd := generatePerfdata(samplestream.Metric.String(), valueNumber, warn, crit)

					// Generate Perfdata from API return
					if !math.IsInf(numberValue, 0) && !math.IsNaN(numberValue) {
						partial.Perfdata.Add(&pd)
					}
				}

				overall.AddSubcheck(partial)
			}
		}

		if len(warnings) != 0 {
			appendum := fmt.Sprintf("HTTP Warnings: %v", strings.Join(warnings, ", "))
			overall.Summary = overall.GetOutput() + appendum
		}
		check.ExitRaw(overall.GetStatus(), overall.GetOutput())
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
