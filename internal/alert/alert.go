package alert

import (
	"fmt"
	"github.com/NETWAYS/go-check"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"strconv"
)

type Alertingrule struct {
	AlertingRule  v1.AlertingRule
	Alert         *v1.Alert
	RecordingRule v1.RecordingRule
}

func (a *Alertingrule) GetStatus() (status int) {
	switch a.AlertingRule.State {
	case string(v1.AlertStateFiring):
		status = check.Critical
	case string(v1.AlertStatePending):
		status = check.Warning
	case string(v1.AlertStateInactive):
		status = check.OK
	default:
		status = check.Unknown
	}

	return status
}

func (a *Alertingrule) GetOutput() (output string) {
	var (
		job      string
		instance string
	)

	if a.Alert == nil {
		return fmt.Sprintf("[%s] is %s",
			a.AlertingRule.Name,
			a.AlertingRule.State)
	} else {
		labels := a.Alert.Labels
		for key, val := range labels {
			switch key {
			case "instance":
				instance = string(val)
			case "job":
				job = string(val)
			}
		}

		val, err := strconv.ParseFloat(a.Alert.Value, 32)
		if err != nil {
			check.ExitError(err)
		}

		output += fmt.Sprintf("[%s] - Job: [%s]", a.AlertingRule.Name, job)

		//if a.AlertingRule.State != "inactive" {
		//	output += fmt.Sprintf(" on Instance: [%s] ", instance)
		//}

		if instance != "" {
			output += fmt.Sprintf(" on Instance: [%s]", instance)
		}

		if a.AlertingRule.State != "inactive" {
			output += fmt.Sprintf(" is %s - value: %.2f",
				a.AlertingRule.State,
				val)
		}

		return output
	}
}
