package alert

import (
	"github.com/prometheus/common/model"
)

// Alert represents an Alertmanager alert
type AlertmanagerAlert struct {
	Annotations struct {
		Description string `json:"description"`
		Summary     string `json:"summary"`
	} `json:"annotations"`
	Status struct {
		// unprocessed, active, suppressed
		State string `json:"state"`
	} `json:"status"`
	Labels model.LabelSet `json:"labels"`
}

func (al *AlertmanagerAlert) GetState() string {
	return al.Status.State
}

func (al *AlertmanagerAlert) GetName() string {
	n, ok := al.Labels["alertname"]

	if ok {
		return string(n)
	}

	return ""
}
