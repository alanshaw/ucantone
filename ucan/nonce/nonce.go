package nonce

import "crypto/rand"

// Generate creates a slice of the given size filled with random bytes.
// Note: the spec recommends at least 12 bytes.
func Generate(size int) []byte {
	out := make([]byte, size)
	rand.Read(out)
	return out
}
