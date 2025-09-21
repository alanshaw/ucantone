package invocation

import (
	"testing"

	"github.com/alanshaw/ucantone/did"
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/common"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	issuer, err := did.Parse("did:key:z6Mkod5Jr3yd5SC7UDueqK4dAAw5xYJYjksy722tA9Boxc4z")
	require.NoError(t, err)

	headerBytes, err := varsig.Encode(common.Ed25519DagCbor)
	require.NoError(t, err)

	model := idm.EnvelopeModel{
		Signature: []byte{},
		SigPayload: idm.SigPayloadModel{
			Header: headerBytes,
			TokenPayload1_0_0_rc1: &idm.TokenPayloadModel1_0_0_rc1{
				Iss: issuer,
			},
		},
	}
	inv := Invocation{model: &model}

	enc, err := Encode(&inv)
	require.NoError(t, err)

	_, err = Decode(enc)
	require.NoError(t, err)
}
