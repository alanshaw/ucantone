package datamodel

import (
	"github.com/alanshaw/ucantone/ipld/datamodel"
)

type ResultModel struct {
	Ok  *datamodel.Any `cborgen:"ok,omitempty"`
	Err *datamodel.Any `cborgen:"error,omitempty"`
}
