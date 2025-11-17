package invocation

import (
	"github.com/alanshaw/ucantone/ipld/datamodel"
)

// UnknownArguments can be used when the arguments for an invocation cannot be
// bound to a known type.
type UnknownArguments = *datamodel.Map
