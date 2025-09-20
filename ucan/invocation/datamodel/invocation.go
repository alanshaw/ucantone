package datamodel

import (
	"github.com/alanshaw/ucantone/did"
	edm "github.com/alanshaw/ucantone/ucan/envelope/datamodel"
)

type InvocationModel struct {
	Header           []byte                    `cborgen:"h"`
	Payload1_0_0_rc1 *InvocationModel1_0_0_rc1 `cborgen:"ucan/inv@1.0.0-rc.1,omitempty"`
}

type InvocationModel1_0_0_rc1 struct {
	Iss did.DID `cborgen:"iss"`
}

type EnvelopeModel edm.EnvelopeModel[InvocationModel]
