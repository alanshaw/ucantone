package execution

import (
	"fmt"

	edm "github.com/alanshaw/ucantone/errors/datamodel"
	"github.com/alanshaw/ucantone/ucan"
)

const HandlerExecutionErrorName = "HandlerExecutionError"

func NewHandlerExecutionError(cmd ucan.Command, cause error) error {
	return edm.ErrorModel{
		ErrorName: HandlerExecutionErrorName,
		Message:   fmt.Errorf("%q handler execution error: %w", cmd, cause).Error(),
	}
}

const InvalidAudienceErrorName = "InvalidAudience"

func NewInvalidAudienceError(expected ucan.Principal, actual ucan.Principal) error {
	return edm.ErrorModel{
		ErrorName: InvalidAudienceErrorName,
		Message:   fmt.Errorf("invalid audience: expected %q, got %q", expected.DID(), actual.DID()).Error(),
	}
}
