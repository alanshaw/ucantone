package datamodel

import (
	"github.com/alanshaw/ucantone/did"
	edm "github.com/alanshaw/ucantone/ucan/envelope/datamodel"
)

type TokenPayloadModel1_0_0_rc1 struct {
	Iss did.DID `cborgen:"iss"`
}

type SigPayloadModel struct {
	Header                []byte                      `cborgen:"h"`
	TokenPayload1_0_0_rc1 *TokenPayloadModel1_0_0_rc1 `cborgen:"ucan/inv@1.0.0-rc.1,omitempty"`
}

type EnvelopeModel edm.EnvelopeModel[SigPayloadModel]
