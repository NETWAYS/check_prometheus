package query

import (
	"fmt"
	"github.com/NETWAYS/go-check"
	"github.com/prometheus/common/model"
)

type ValType struct {
	Scalar    model.Scalar
	Vector    model.Vector
	Matrix    model.Matrix
	String    model.String
	Critical  check.Threshold
	Warning   check.Threshold
	ValueType model.ValueType
}

func (v *ValType) GetValType(queryResult model.Value) {
	switch queryResult.Type() {
	case model.ValScalar:
		v.ValueType = model.ValScalar
	case model.ValVector:
		v.ValueType = model.ValVector
	case model.ValMatrix:
		v.ValueType = model.ValMatrix
	case model.ValString:
		v.ValueType = model.ValString
	case model.ValNone:
		v.ValueType = model.ValNone
	default:
		check.ExitError(fmt.Errorf("unkown model type. Please examine the query in the frontend"))
	}
}

func (v *ValType) GetStatus() (status int) {
	return status
}

// \_[CRITICAL] [HighResultLatency] - Job: [prometheus] on Instance: [localhost:9090] is firing
func (v *ValType) GetOuput() (output string) {
	return output
}
