package main

import (
	"fmt"
	"github.com/NETWAYS/check_prometheus/cmd"
)

// nolint: gochecknoglobals
var (
	version = "0.1.0"
	commit  = ""
	date    = ""
)

func main() {
	cmd.Execute(buildVersion())
}

//goland:noinspection GoBoolExpressions
func buildVersion() string {
	result := version

	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}

	if date != "" {
		result = fmt.Sprintf("%s\ndate: %s", result, date)
	}

	return result
}
