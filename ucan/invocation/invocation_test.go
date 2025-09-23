package invocation

import (
	"bytes"
	"testing"

	"github.com/alanshaw/ucantone/testing/helpers"
	"github.com/alanshaw/ucantone/ucan/crypto/signature"
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/common"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	issuer := helpers.RandomSigner(t)
	subject := helpers.RandomDID(t)

	headerBytes, err := varsig.Encode(common.Ed25519DagCbor)
	require.NoError(t, err)

	sigPayload := idm.SigPayloadModel{
		Header: headerBytes,
		TokenPayload1_0_0_rc1: &idm.TokenPayloadModel1_0_0_rc1{
			Iss: issuer.DID(),
			Sub: subject,
		},
	}

	var buf bytes.Buffer
	err = sigPayload.MarshalCBOR(&buf)
	require.NoError(t, err)

	model := idm.EnvelopeModel{
		Signature:  issuer.Sign(buf.Bytes()),
		SigPayload: sigPayload,
	}
	sig := signature.NewSignature(common.Ed25519DagCbor, model.Signature)
	inv := Invocation{sig: sig, model: &model}

	enc, err := Encode(&inv)
	require.NoError(t, err)

	_, err = Decode(enc)
	require.NoError(t, err)
}
