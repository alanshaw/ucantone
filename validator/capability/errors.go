package capability

import (
	"fmt"

	edm "github.com/alanshaw/ucantone/errors/datamodel"
	"github.com/alanshaw/ucantone/ucan"
)

const MalformedArgumentsErrorName = "MalformedArguments"

func NewMalformedArgumentsError(cmd ucan.Command, cause error) edm.ErrorModel {
	return edm.ErrorModel{
		ErrorName: MalformedArgumentsErrorName,
		Message:   fmt.Sprintf("malformed arguments for command %s: %s", cmd, cause.Error()),
	}
}
