package alert

import (
	"fmt"
	"github.com/NETWAYS/go-check"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Alertingrule struct {
	Alertingrule v1.AlertingRule
	Alert        *v1.Alert
}

func (a *Alertingrule) GetStatus() (status int) {
	switch a.Alertingrule.State {
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
			a.Alertingrule.Name,
			a.Alertingrule.State)
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

		return fmt.Sprintf("[%s] - Job: [%s] on Instance: [%s] is %s",
			a.Alertingrule.Name,
			job,
			instance,
			a.Alertingrule.State)
	}
}
