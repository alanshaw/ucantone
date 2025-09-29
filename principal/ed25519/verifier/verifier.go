package verifier

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"strings"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/principal"
	"github.com/alanshaw/ucantone/principal/multiformat"
	vsed "github.com/alanshaw/ucantone/varsig/algorithm/ed25519"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-varint"
)

const Code = 0xed
const SignatureCode = vsed.Code

var publicTagSize = varint.UvarintSize(Code)

const keySize = ed25519.PublicKeySize

var size = publicTagSize + keySize

func Parse(str string) (Ed25519Verifier, error) {
	if !strings.HasPrefix(str, did.KeyPrefix) {
		return nil, fmt.Errorf("must start with '%s'", did.KeyPrefix)
	}
	code, bytes, err := multibase.Decode(str[len(did.KeyPrefix):])
	if err != nil {
		return nil, err
	}
	if code != multibase.Base58BTC {
		return nil, errors.New("not Base58BTC encoded")
	}
	return Decode(bytes)
}

func Decode(b []byte) (Ed25519Verifier, error) {
	if len(b) != size {
		return nil, fmt.Errorf("invalid length: %d wanted: %d", len(b), size)
	}
	code, _, err := varint.FromUvarint(b)
	if err != nil {
		return nil, fmt.Errorf("reading uvarint: %w", err)
	}
	if code != Code {
		return nil, fmt.Errorf("invalid public key codec: 0x%02x, expected: 0x%02x", code, Code)
	}
	v := make(Ed25519Verifier, size)
	copy(v, b)
	return v, nil
}

// FromRaw takes raw ed25519 public key bytes and tags with the ed25519 verifier
// multiformat code, returning an ed25519 verifier.
func FromRaw(b []byte) (Ed25519Verifier, error) {
	if len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid length: %d wanted: %d", len(b), ed25519.PublicKeySize)
	}
	return Ed25519Verifier(multiformat.TagWith(Code, b)), nil
}

type Ed25519Verifier []byte

var _ principal.Verifier = (Ed25519Verifier)(nil)

func (v Ed25519Verifier) Code() uint64 {
	return Code
}

func (v Ed25519Verifier) Verify(msg []byte, sig []byte) bool {
	return ed25519.Verify(ed25519.PublicKey(v[publicTagSize:]), msg, sig)
}

func (v Ed25519Verifier) DID() did.DID {
	b58key, _ := multibase.Encode(multibase.Base58BTC, v)
	id, _ := did.Parse(did.KeyPrefix + b58key)
	return id
}

// Bytes returns the public key bytes with multiformat prefix varint.
func (v Ed25519Verifier) Bytes() []byte {
	return v
}

// Raw encodes the bytes of the public key without multiformats tags.
func (s Ed25519Verifier) Raw() []byte {
	k := make(ed25519.PublicKey, ed25519.PublicKeySize)
	copy(k, s[publicTagSize:])
	return k
}
