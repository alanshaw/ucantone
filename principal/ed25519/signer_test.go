package ed25519_test

import (
	"crypto/ed25519"
	"fmt"
	"testing"

	ed "github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/principal/signer"
	"github.com/stretchr/testify/require"
)

func TestGenerateEncodeDecode(t *testing.T) {
	s0, err := ed.Generate()
	if err != nil {
		t.Fatalf("generating Ed25519 key: %v", err)
	}

	fmt.Println(s0.DID().String())

	s1, err := ed.Decode(s0.Bytes())
	if err != nil {
		t.Fatalf("decoding Ed25519 key: %v", err)
	}

	fmt.Println(s1.DID().String())

	if s0.DID().String() != s1.DID().String() {
		t.Fatalf("public key mismatch: %s != %s", s0.DID().String(), s1.DID().String())
	}
}

func TestGenerateFormatParse(t *testing.T) {
	s0, err := ed.Generate()
	if err != nil {
		t.Fatalf("generating Ed25519 key: %v", err)
	}

	fmt.Println(s0.DID().String())

	str, err := signer.Format(s0)
	if err != nil {
		t.Fatalf("formatting Ed25519 key: %v", err)
	}

	fmt.Println(str)

	s1, err := ed.Parse(str)
	if err != nil {
		t.Fatalf("parsing Ed25519 key: %v", err)
	}

	fmt.Println(s1.DID().String())

	if s0.DID().String() != s1.DID().String() {
		t.Fatalf("public key mismatch: %s != %s", s0.DID().String(), s1.DID().String())
	}
}

func TestVerify(t *testing.T) {
	s0, err := ed.Generate()
	if err != nil {
		t.Fatalf("generating Ed25519 key: %v", err)
	}

	msg := []byte("testy")
	sig := s0.Sign(msg)

	res := s0.Verifier().Verify(msg, sig)
	if res != true {
		t.Fatalf("verify failed")
	}
}

func TestSignerRaw(t *testing.T) {
	s, err := ed.Generate()
	require.NoError(t, err)

	msg := []byte{1, 2, 3}
	raw := s.Raw()
	sk := ed25519.NewKeyFromSeed(raw)
	sig := ed25519.Sign(sk, msg)

	require.Equal(t, s.Sign(msg), sig)
}

func TestFromRaw(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		_, priv, err := ed25519.GenerateKey(nil)
		require.NoError(t, err)

		s, err := ed.FromRaw(priv[:ed25519.SeedSize])
		require.NoError(t, err)

		require.Equal(t, []byte(priv[:ed25519.SeedSize]), s.Raw())
	})

	t.Run("invalid length", func(t *testing.T) {
		_, err := ed.FromRaw([]byte{})
		require.Error(t, err)
		require.ErrorContains(t, err, "invalid length")
	})
}
