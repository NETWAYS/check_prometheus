package query

import (
	"github.com/NETWAYS/go-check"
	"github.com/prometheus/common/model"
)

type ValType struct {
	Scalar   model.Scalar
	Vector   model.Vector
	Matrix   model.Matrix
	String   model.String
	Critical check.Threshold
	Warning  check.Threshold
}

func (v *ValType) GetStatus() (status int) {

	return status
}
