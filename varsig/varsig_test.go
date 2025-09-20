package varsig_test

import (
	"encoding/base64"
	"testing"

	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/algoithm/ed25519"
	"github.com/alanshaw/ucantone/varsig/payload/dagcbor"
	"github.com/stretchr/testify/require"
)

func TestVarsig(t *testing.T) {
	varsig.RegisterSignatureAlgorithm(ed25519.NewCodec())
	varsig.RegisterPayloadEncoding(dagcbor.NewCodec())

	expectHeader := varsig.NewHeader(ed25519.New(), dagcbor.New())

	data, err := varsig.Encode(expectHeader)
	require.NoError(t, err)

	t.Log("Encoded (base64):")
	t.Logf("\t%s", base64.RawStdEncoding.EncodeToString(data))

	header, err := varsig.Decode(data)
	require.NoError(t, err)

	require.Equal(t, expectHeader.Version(), header.Version())
	require.Equal(t, expectHeader.SignatureAlgorithm().Code(), header.SignatureAlgorithm().Code())

	sigAlgo, ok := header.SignatureAlgorithm().(ed25519.SignatureAlgorithm)
	require.True(t, ok)
	require.Equal(t, expectHeader.SignatureAlgorithm().Code(), sigAlgo.Code())
	require.Equal(t, expectHeader.SignatureAlgorithm().Curve(), sigAlgo.Curve())
	require.Equal(t, expectHeader.SignatureAlgorithm().HashAlgorithm(), sigAlgo.HashAlgorithm())

	t.Log("Signature Algorithm:")
	t.Logf("\tCode:\t0x%02x", sigAlgo.Code())
	t.Logf("\tCurve:\t0x%02x", sigAlgo.Curve())
	t.Logf("\tHash:\t0x%02x", sigAlgo.HashAlgorithm())

	payloadEnc, ok := header.PayloadEncoding().(dagcbor.PayloadEncoding)
	require.True(t, ok)
	require.Equal(t, expectHeader.PayloadEncoding().Code(), payloadEnc.Code())

	t.Log("Payload Encoing:")
	t.Logf("\tCode:\t0x%02x", payloadEnc.Code())
}
