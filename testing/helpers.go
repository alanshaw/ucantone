package helpers

import (
	crand "crypto/rand"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
)

// Must takes return values from a function and returns the non-error one. If
// the error value is non-nil then it panics.
func Must[T any](val T, err error) func(t *testing.T) T {
	return func(t *testing.T) T {
		require.NoError(t, err)
		return val
	}
}

func RandomBytes(t *testing.T, size int) []byte {
	t.Helper()
	bytes := make([]byte, size)
	_, err := crand.Read(bytes)
	require.NoError(t, err)
	return bytes
}

func RandomCID(t *testing.T) cid.Cid {
	t.Helper()
	bytes := RandomBytes(t, 10)
	c, _ := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   multihash.SHA2_256,
		MhLength: -1,
	}.Sum(bytes)
	return c
}

func RandomDigest(t *testing.T) multihash.Multihash {
	t.Helper()
	return RandomCID(t).Hash()
}
