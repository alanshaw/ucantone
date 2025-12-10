package executor

import (
	"fmt"

	edm "github.com/alanshaw/ucantone/errors/datamodel"
	"github.com/alanshaw/ucantone/ucan"
)

const HandlerNotFoundErrorName = "HandlerNotFound"

func NewHandlerNotFoundError(cmd ucan.Command) error {
	return edm.ErrorModel{
		ErrorName: HandlerNotFoundErrorName,
		Message:   fmt.Sprintf("handler not found: %s", cmd),
	}
}

const HandlerExecutionErrorName = "HandlerExecutionError"

func NewHandlerExecutionError(cmd ucan.Command, cause error) error {
	return edm.ErrorModel{
		ErrorName: HandlerExecutionErrorName,
		Message:   fmt.Errorf("%s handler execution error: %w", cmd, cause).Error(),
	}
}
