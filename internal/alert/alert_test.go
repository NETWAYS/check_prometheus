package alert

import (
	"github.com/NETWAYS/go-check"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"testing"
	"time"
)

func TestGetStatus(t *testing.T) {

	testTime := time.Now()

	ar := v1.AlertingRule{
		Alerts: []*v1.Alert{
			{
				ActiveAt: testTime.UTC(),
				Annotations: model.LabelSet{
					"summary": "High request latency",
				},
				Labels: model.LabelSet{
					"alertname": "HighRequestLatency",
					"severity":  "page",
				},
				State: v1.AlertStateFiring,
				Value: "1e+00",
			},
		},
		Annotations: model.LabelSet{
			"summary": "High request latency",
		},
		Labels: model.LabelSet{
			"severity": "page",
		},
		Duration:       600,
		Health:         v1.RuleHealthGood,
		Name:           "HighRequestLatency",
		Query:          "job:request_latency_seconds:mean5m{job=\"myjob\"} > 0.5",
		LastError:      "",
		EvaluationTime: 0.5,
		LastEvaluation: time.Date(2020, 5, 18, 15, 52, 53, 450311300, time.UTC),
		State:          "firing",
	}

	rr := v1.RecordingRule{
		Health:         v1.RuleHealthGood,
		Name:           "job:http_inprogress_requests:sum",
		Query:          "sum(http_inprogress_requests) by (job)",
		LastError:      "",
		EvaluationTime: 0.3,
		LastEvaluation: time.Date(2020, 5, 18, 15, 52, 53, 450311300, time.UTC),
	}

	r := Rule{
		AlertingRule:  ar,
		Alert:         ar.Alerts[0],
		RecordingRule: rr,
	}

	actual := r.GetStatus()
	if actual != check.Critical {
		t.Error("\nActual: ", actual, "\nExpected: ", check.Critical)
	}

	r.AlertingRule.State = "pending"
	actual = r.GetStatus()
	if actual != check.Warning {
		t.Error("\nActual: ", actual, "\nExpected: ", check.Warning)
	}

}

func TestGetOutput(t *testing.T) {

	testTime := time.Now()

	ar := v1.AlertingRule{
		Alerts: []*v1.Alert{
			{
				ActiveAt: testTime.UTC(),
				Annotations: model.LabelSet{
					"summary": "High request latency",
				},
				Labels: model.LabelSet{
					"alertname": "HighRequestLatency",
					"instance":  "foo",
					"job":       "bar",
				},
				State: v1.AlertStateFiring,
				Value: "1e+00",
			},
		},
		Annotations: model.LabelSet{
			"summary": "High request latency",
		},
		Labels: model.LabelSet{
			"severity": "page",
		},
		Duration:       600,
		Health:         v1.RuleHealthGood,
		Name:           "HighRequestLatency",
		Query:          "job:request_latency_seconds:mean5m{job=\"myjob\"} > 0.5",
		LastError:      "",
		EvaluationTime: 0.5,
		LastEvaluation: time.Date(2020, 5, 18, 15, 52, 53, 450311300, time.UTC),
		State:          "firing",
	}

	rr := v1.RecordingRule{
		Health:         v1.RuleHealthGood,
		Name:           "job:http_inprogress_requests:sum",
		Query:          "sum(http_inprogress_requests) by (job)",
		LastError:      "",
		EvaluationTime: 0.3,
		LastEvaluation: time.Date(2020, 5, 18, 15, 52, 53, 450311300, time.UTC),
	}

	r := Rule{
		AlertingRule:  ar,
		Alert:         ar.Alerts[0],
		RecordingRule: rr,
	}

	var expected string

	expected = "[HighRequestLatency] - Job: [bar] on Instance: [foo] is firing - value: 1.00"
	if r.GetOutput() != expected {
		t.Error("\nActual: ", r.GetOutput(), "\nExpected: ", expected)
	}

	r.AlertingRule.Alerts[0].Labels = model.LabelSet{
		"alertname": "HighRequestLatency",
	}

	expected = "[HighRequestLatency] is firing - value: 1.00"
	if r.GetOutput() != expected {
		t.Error("\nActual: ", r.GetOutput(), "\nExpected: ", expected)
	}

	r.AlertingRule.State = "inactive"

	expected = "[HighRequestLatency] is inactive - value: 1.00"
	if r.GetOutput() != expected {
		t.Error("\nActual: ", r.GetOutput(), "\nExpected: ", expected)
	}
}
