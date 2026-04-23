package secp256k1_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"testing"

	secp256k1 "github.com/alanshaw/ucantone/principal/secp256k1"
	"github.com/alanshaw/ucantone/principal/signer"
	"github.com/ethereum/go-ethereum/crypto"
	eth_secp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

func TestGenerateEncodeDecode(t *testing.T) {
	s0, err := secp256k1.Generate()
	require.NoError(t, err)

	t.Log(s0.DID().String())

	s1, err := secp256k1.Decode(s0.Bytes())
	require.NoError(t, err)

	t.Log(s1.DID().String())
	require.Equal(t, s0.DID(), s1.DID(), "public key mismatch")
}

func TestGenerateFormatParse(t *testing.T) {
	s0, err := secp256k1.Generate()
	require.NoError(t, err)

	t.Log(s0.DID().String())

	str := signer.Format(s0)
	t.Log(str)

	s1, err := secp256k1.Parse(str)
	require.NoError(t, err)

	t.Log(s1.DID().String())
	require.Equal(t, s0.DID(), s1.DID(), "public key mismatch")
}

func TestVerify(t *testing.T) {
	s, err := secp256k1.Generate()
	require.NoError(t, err)

	msg := []byte("testy")
	sig := s.Sign(msg)

	res := s.Verifier().Verify(msg, sig)
	require.True(t, res)
}

func TestSignerRaw(t *testing.T) {
	s, err := secp256k1.Generate()
	require.NoError(t, err)

	msg := []byte{1, 2, 3}
	hash := sha256.New()
	hash.Write(msg)
	raw := s.Raw()
	sig, err := eth_secp256k1.Sign(hash.Sum(nil), raw)
	require.NoError(t, err)

	require.Equal(t, s.Sign(msg), sig[:crypto.RecoveryIDOffset])
}

func TestFromRaw(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		priv, err := ecdsa.GenerateKey(eth_secp256k1.S256(), rand.Reader)
		require.NoError(t, err)

		s, err := secp256k1.FromRaw(priv.D.Bytes())
		require.NoError(t, err)

		require.Equal(t, priv.D.Bytes(), s.Raw())
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := secp256k1.FromRaw([]byte{})
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid length")
	})
}
