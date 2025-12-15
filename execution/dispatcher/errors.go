package dispatcher

import (
	"fmt"

	edm "github.com/alanshaw/ucantone/errors/datamodel"
	"github.com/alanshaw/ucantone/ucan"
)

const HandlerNotFoundErrorName = "HandlerNotFound"

func NewHandlerNotFoundError(cmd ucan.Command) error {
	return edm.ErrorModel{
		ErrorName: HandlerNotFoundErrorName,
		Message:   fmt.Sprintf("handler not found: %q", cmd),
	}
}
