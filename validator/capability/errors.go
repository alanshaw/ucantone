package capability

import (
	"fmt"

	"github.com/alanshaw/ucantone/ucan"
	vdm "github.com/alanshaw/ucantone/validator/errors/datamodel"
)

const MalformedArgumentsErrorName = "MalformedArguments"

func NewMalformedArgumentsError(cmd ucan.Command, cause error) vdm.ErrorModel {
	return vdm.ErrorModel{
		ErrorName: MalformedArgumentsErrorName,
		Message:   fmt.Sprintf("malformed arguments for command %s: %s", cmd, cause.Error()),
	}
}
