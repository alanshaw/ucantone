package errors

import (
	edm "github.com/alanshaw/ucantone/errors/datamodel"
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
