package verifier_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"

	"github.com/alanshaw/ucantone/principal/secp256k1/verifier"
	eth_secp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	str := "did:key:zQ3shokFvN6Ggnq5j6G76527464y7n7y767y767y767y767y7"
	v, err := verifier.Parse(str)
	require.NoError(t, err)
	require.Equal(t, str, v.DID().String())
}

func TestFromRaw(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		priv, err := ecdsa.GenerateKey(eth_secp256k1.S256(), rand.Reader)
		require.NoError(t, err)

		pub := eth_secp256k1.CompressPubkey(priv.PublicKey.X, priv.PublicKey.Y)
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
