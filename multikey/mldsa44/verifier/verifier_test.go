package verifier_test

// This test package deliberately avoids importing the mldsa44 signer package:
// it registers the ML-DSA-44 Varsig algorithm too, which would mask a missing
// registration in a verifier-only import path (see TestVarsigAlgorithmRegistered).

import (
	"crypto/mldsa"
	"testing"

	"github.com/fil-forge/ucantone/multikey"
	"github.com/fil-forge/ucantone/multikey/mldsa44/verifier"
	"github.com/fil-forge/ucantone/varsig"
	"github.com/stretchr/testify/require"
)

// TestParseKeyDID round-trips a `did:key` string through [verifier.ParseKeyDID].
// ML-DSA-44 public keys are 1312 bytes, so rather than hardcode a DID we derive
// one from a freshly generated key.
func TestParseKeyDID(t *testing.T) {
	sk, err := mldsa.GenerateKey(mldsa.MLDSA44())
	require.NoError(t, err)

	v, err := verifier.FromRaw(sk.PublicKey().Bytes())
	require.NoError(t, err)

	str := v.KeyDID().String()
	v2, err := verifier.ParseKeyDID(str)
	require.NoError(t, err)
	require.Equal(t, str, v2.KeyDID().String())
}

func TestDecode(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		sk, err := mldsa.GenerateKey(mldsa.MLDSA44())
		require.NoError(t, err)

		v, err := verifier.FromRaw(sk.PublicKey().Bytes())
		require.NoError(t, err)

		v2, err := verifier.Decode(v.Bytes())
		require.NoError(t, err)
		require.Equal(t, v, v2)
	})
}

func TestFromRaw(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		sk, err := mldsa.GenerateKey(mldsa.MLDSA44())
		require.NoError(t, err)

		pub := sk.PublicKey().Bytes()
		v, err := verifier.FromRaw(pub)
		require.NoError(t, err)

		require.Equal(t, pub, v.Raw())
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := verifier.FromRaw([]byte{})
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid length")
	})
}

// TestParseAndVerify covers the `did:key` -> [multikey.Parse] -> verify path,
// confirming the package's Decoder is registered with the multikey registry via
// its init and that a signature verifies against the parsed verifier.
func TestParseAndVerify(t *testing.T) {
	sk, err := mldsa.GenerateKey(mldsa.MLDSA44())
	require.NoError(t, err)

	msg := []byte("testy")
	sig, err := sk.SignDeterministic(msg, nil)
	require.NoError(t, err)

	v0, err := verifier.FromRaw(sk.PublicKey().Bytes())
	require.NoError(t, err)

	v, err := multikey.Parse(v0.KeyDID().Identifier())
	require.NoError(t, err)
	require.True(t, v.Verify(msg, sig))
}

// TestVarsigAlgorithmRegistered confirms that importing this package alone is
// enough for [varsig.Decode] to recognise the ML-DSA-44 algorithm scheme, i.e.
// the blank import of varsig/algorithm/mldsa runs its init registration in
// verifier-only binaries. The header is hardcoded so the test doesn't import
// the algorithm package (which would register it itself).
func TestVarsigAlgorithmRegistered(t *testing.T) {
	// varsig prefix (0x34), version (0x01), ML-DSA-44 code (0x1210 as varint),
	// DAG-CBOR payload encoding (0x71).
	header := []byte{0x34, 0x01, 0x90, 0x24, 0x71}

	v, n, err := varsig.Decode(header)
	require.NoError(t, err)
	require.Equal(t, len(header), n)
	require.Equal(t, varsig.DagCbor, v.PayloadEncoding())

	// The decoded algorithm should re-encode to the same header segment.
	algBytes, err := v.SignatureAlgorithm().Encode()
	require.NoError(t, err)
	require.Equal(t, []byte{0x90, 0x24}, algBytes)
}
