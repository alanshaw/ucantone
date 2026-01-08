package errors

import (
	"errors"

	edm "github.com/alanshaw/ucantone/errors/datamodel"
)

var (
	Is = errors.Is
	As = errors.As
)

type Named interface {
	error
	Name() string
}

func New(name, message string) error {
	return edm.ErrorModel{
		ErrorName: name,
		Message:   message,
	}
}
