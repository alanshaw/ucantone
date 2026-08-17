package receipt_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	cbg "github.com/whyrusleeping/cbor-gen"

	"github.com/fil-forge/ucantone/ipld/datamodel"
	rsdm "github.com/fil-forge/ucantone/result/datamodel"
	"github.com/fil-forge/ucantone/testutil"
	"github.com/fil-forge/ucantone/ucan"
	"github.com/fil-forge/ucantone/ucan/invocation"
	"github.com/fil-forge/ucantone/ucan/receipt"
	rdm "github.com/fil-forge/ucantone/ucan/receipt/datamodel"
	"github.com/fil-forge/ucantone/ucan/token"
)

func TestIssueOK(t *testing.T) {
	executor := testutil.RandomIssuer(t)
	ran := testutil.RandomCID(t)

	ok := cbg.CborInt(42)
	initial, err := receipt.IssueOK(executor, ran, &ok)
	require.NoError(t, err)

	encoded, err := receipt.Encode(initial)
	require.NoError(t, err)

	decoded, err := receipt.Decode(encoded)
	require.NoError(t, err)

	require.Equal(t, executor.DID(), decoded.Issuer())
	require.Equal(t, ran, decoded.Ran())
	require.NotEmpty(t, decoded.Nonce())

	okBytes, errBytes := decoded.Out().Unpack()
	require.Nil(t, errBytes)

	var got cbg.CborInt
	require.NoError(t, got.UnmarshalCBOR(bytes.NewReader(okBytes)))
	require.Equal(t, cbg.CborInt(42), got)
}

func TestIssueErr(t *testing.T) {
	executor := testutil.RandomIssuer(t)
	ran := testutil.RandomCID(t)

	errVal := cbg.CborInt(7)
	initial, err := receipt.IssueErr(executor, ran, &errVal)
	require.NoError(t, err)

	decoded, err := receipt.Decode(testutil.Must(receipt.Encode(initial))(t))
	require.NoError(t, err)

	okBytes, errBytes := decoded.Out().Unpack()
	require.Nil(t, okBytes)

	var got cbg.CborInt
	require.NoError(t, got.UnmarshalCBOR(bytes.NewReader(errBytes)))
	require.Equal(t, cbg.CborInt(7), got)
}

func TestOptions(t *testing.T) {
	executor := testutil.RandomIssuer(t)
	ran := testutil.RandomCID(t)

	t.Run("custom issued at", func(t *testing.T) {
		ok := cbg.CborInt(1)
		now := ucan.Now()
		rcpt, err := receipt.IssueOK(executor, ran, &ok, receipt.WithIssuedAt(now))
		require.NoError(t, err)
		require.NotNil(t, rcpt.IssuedAt())
		require.Equal(t, now, *rcpt.IssuedAt())
	})

	t.Run("no issued at by default", func(t *testing.T) {
		ok := cbg.CborInt(1)
		rcpt, err := receipt.IssueOK(executor, ran, &ok)
		require.NoError(t, err)
		require.Nil(t, rcpt.IssuedAt())

		decoded, err := receipt.Decode(testutil.Must(receipt.Encode(rcpt))(t))
		require.NoError(t, err)
		require.Nil(t, decoded.IssuedAt())
	})

	t.Run("custom nonce", func(t *testing.T) {
		ok := cbg.CborInt(1)
		nonce := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		rcpt, err := receipt.IssueOK(executor, ran, &ok, receipt.WithNonce(nonce))
		require.NoError(t, err)
		require.Equal(t, nonce, rcpt.Nonce())
	})

	t.Run("no nonce", func(t *testing.T) {
		ok := cbg.CborInt(1)
		rcpt, err := receipt.IssueOK(executor, ran, &ok, receipt.WithNoNonce())
		require.NoError(t, err)
		require.Empty(t, rcpt.Nonce())
	})

	t.Run("WithExpiration", func(t *testing.T) {
		ok := cbg.CborInt(1)
		exp := ucan.Now() + 3600
		rcpt, err := receipt.IssueOK(executor, ran, &ok, receipt.WithExpiration(exp))
		require.NoError(t, err)
		require.NotNil(t, rcpt.Expiration())
		require.Equal(t, exp, *rcpt.Expiration())

		decoded, err := receipt.Decode(testutil.Must(receipt.Encode(rcpt))(t))
		require.NoError(t, err)
		require.NotNil(t, decoded.Expiration())
		require.Equal(t, exp, *decoded.Expiration())
	})

	t.Run("DefaultNoExpiration", func(t *testing.T) {
		ok := cbg.CborInt(1)
		rcpt, err := receipt.IssueOK(executor, ran, &ok)
		require.NoError(t, err)
		require.Nil(t, rcpt.Expiration())
	})
}

// TestRejectsAudience asserts that an invocation carrying an audience does
// not decode as a receipt: the spec requires aud to be omitted.
func TestRejectsAudience(t *testing.T) {
	executor := testutil.RandomIssuer(t)
	ran := testutil.RandomCID(t)

	ok := cbg.CborInt(1)
	var buf bytes.Buffer
	require.NoError(t, ok.MarshalCBOR(&buf))
	raw := datamodel.NewRaw(buf.Bytes())

	inv, err := invocation.Invoke(
		executor,
		executor.DID(),
		receipt.Command,
		&rdm.ArgsModel{Ran: ran, Out: rsdm.ResultModel{Ok: &raw}},
		invocation.WithAudience(executor.DID()),
	)
	require.NoError(t, err)

	_, err = receipt.Decode(inv.Bytes())
	require.ErrorContains(t, err, "audience must be omitted")
}

// TestNotInvocation asserts a Receipt does NOT satisfy ucan.Invocation. The
// wire format is still an invocation, but the Go type is deliberately its own
// thing so callers can't accidentally pass a receipt where an invocation is
// expected.
func TestNotInvocation(t *testing.T) {
	executor := testutil.RandomIssuer(t)
	ran := testutil.RandomCID(t)
	ok := cbg.CborInt(1)
	rcpt, err := receipt.IssueOK(executor, ran, &ok)
	require.NoError(t, err)

	var _ ucan.Receipt = rcpt
	_, isInv := any(rcpt).(ucan.Invocation)
	require.False(t, isInv, "Receipt must not be assignable to ucan.Invocation")
}

// TestWireFormatIsInvocation locks in the invariant the receipt API rests on:
// although Receipt is its own Go type, on the wire a receipt is byte-identical
// to a /ucan/assert/receipt invocation. The exact bytes produced by
// receipt.Encode MUST decode cleanly via invocation.Decode, with Ran/Out
// recoverable from the invocation's args. If the receipt encoding ever drifts
// from the invocation encoding, this test fails.
func TestWireFormatIsInvocation(t *testing.T) {
	executor := testutil.RandomIssuer(t)
	ran := testutil.RandomCID(t)
	ok := cbg.CborInt(42)

	rcpt, err := receipt.IssueOK(executor, ran, &ok)
	require.NoError(t, err)

	encoded, err := receipt.Encode(rcpt)
	require.NoError(t, err)

	// The same bytes must decode as an invocation...
	inv, err := invocation.Decode(encoded)
	require.NoError(t, err, "receipt bytes must be a valid invocation on the wire")

	// ...and the invocation view must re-encode to byte-identical output, so
	// the two views are not just compatible but the same wire object.
	reencoded, err := invocation.Encode(inv)
	require.NoError(t, err)
	require.Equal(t, encoded, reencoded, "invocation view must re-encode verbatim")

	// It is the /ucan/assert/receipt invocation shape, with the executor as
	// issuer and subject, and the audience omitted — implicitly the subject,
	// per the invocation spec (ucan-wg/receipt#1).
	require.Equal(t, receipt.Command, inv.Command())
	require.Equal(t, executor.DID(), inv.Issuer())
	require.Equal(t, executor.DID(), inv.Subject())
	require.False(t, inv.Audience().Defined(), "receipt audience must be omitted")

	// Receipts are permanent assertions by default (exp: null).
	require.Nil(t, inv.Expiration())

	// The receipt and invocation views agree on identity and signed bytes.
	require.Equal(t, rcpt.Link(), inv.Link(), "receipt and invocation views share one CID")
	require.Equal(t, rcpt.SignedBytes(), inv.SignedBytes())

	// The bytes verify as a signed invocation — not merely parse as one.
	verified := token.VerifySignature(inv, executor.Verifier())
	require.True(t, verified, "receipt bytes must verify through the invocation path")

	// Ran/Out are carried in the invocation's args, value intact.
	var args rdm.ArgsModel
	require.NoError(t, args.UnmarshalCBOR(bytes.NewReader(inv.ArgumentsBytes())))
	require.Equal(t, ran, args.Ran)
	require.NotNil(t, args.Out.Ok)
	var gotOut cbg.CborInt
	require.NoError(t, gotOut.UnmarshalCBOR(bytes.NewReader(args.Out.Ok.Bytes())))
	require.Equal(t, cbg.CborInt(42), gotOut)

	// And the receipt view round-trips back from those same bytes.
	redecoded, err := receipt.Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, rcpt.Link(), redecoded.Link())
	require.Equal(t, rcpt.Ran(), redecoded.Ran())
}
