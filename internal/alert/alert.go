package alert

import (
	"fmt"
	"github.com/NETWAYS/go-check"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"strconv"
	"strings"
)

// Internal representation of Prometheus Rules
// Alert attribute will be used when iterating over multiple AlertingRules
type Rule struct {
	AlertingRule v1.AlertingRule
	Alert        *v1.Alert
}

func FlattenRules(groups []v1.RuleGroup) []Rule {
	// Flattens a list of RuleGroup containing a list of Rules into
	// a list of internal Alertingrules
	var l int
	// Set initial capacity to reduce memory allocations
	for _, grp := range groups {
		l = l + len(grp.Rules)
	}

	rules := make([]Rule, 0, l)

	var r Rule

	for _, grp := range groups {
		for _, rl := range grp.Rules {
			// For now we only care about AlertingRules
			// since RecodingRules can simply be queried
			if _, ok := rl.(v1.AlertingRule); ok {
				r.AlertingRule = rl.(v1.AlertingRule)
				rules = append(rules, r)
			}
		}
	}

	return rules
}

func (a *Rule) GetStatus() (status int) {
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

func (a *Rule) GetOutput() (output string) {
	if a.Alert == nil {
		return fmt.Sprintf("[%s] is %s",
			a.AlertingRule.Name,
			a.AlertingRule.State)
	}

	var (
		value float64
		v     model.LabelValue
		ok    bool
		out   strings.Builder
	)

	// Base Output
	out.WriteString(fmt.Sprintf("[%s]", a.AlertingRule.Name))

	// Add job if available
	v, ok = a.Alert.Labels["job"]
	if ok {
		out.WriteString(fmt.Sprintf(" - Job: [%s]", string(v)))
	}

	// Add instance if available
	v, ok = a.Alert.Labels["instance"]
	if ok {
		out.WriteString(fmt.Sprintf(" on Instance: [%s]", string(v)))
	}

	// Add current value to output
	value, _ = strconv.ParseFloat(a.Alert.Value, 32)
	out.WriteString(fmt.Sprintf(" is %s - value: %.2f", a.AlertingRule.State, value))

	return out.String()
}
