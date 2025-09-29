package ed25519

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/principal"
	"github.com/alanshaw/ucantone/principal/ed25519/verifier"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-varint"
)

const Code = 0x1300

const SignatureCode = verifier.SignatureCode

var tagSize = varint.UvarintSize(Code)

// Go ed25519 private key size is private + public. Go refers to the private key
// bytes as the "seed".
const keySize = ed25519.SeedSize

var size = tagSize + keySize

func Generate() (Ed25519Signer, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating Ed25519 key: %w", err)
	}
	s := make(Ed25519Signer, size)
	varint.PutUvarint(s, Code)
	copy(s[tagSize:], priv)
	return s, nil
}

// Parse parses a multibase encoded string containing a ed25519 signer
// multiformat varint (0x1300) + 32 byte ed25519 private key
func Parse(str string) (Ed25519Signer, error) {
	_, bytes, err := multibase.Decode(str)
	if err != nil {
		return nil, fmt.Errorf("decoding multibase string: %w", err)
	}
	return Decode(bytes)
}

// Decode decodes a buffer of an ed25519 signer multiformat varint (0x1300) + 32
// byte ed25519 private key.
func Decode(b []byte) (Ed25519Signer, error) {
	if len(b) != size {
		return nil, fmt.Errorf("invalid length: %d wanted: %d", len(b), size)
	}

	skc, _, err := varint.FromUvarint(b)
	if err != nil {
		return nil, fmt.Errorf("reading private key uvarint: %w", err)
	}
	if skc != Code {
		return nil, fmt.Errorf("invalid private key codec: 0x%02x, expected: 0x%02x", skc, Code)
	}

	s := make(Ed25519Signer, size)
	copy(s, b)

	return s, nil
}

// FromRaw takes raw 32 byte ed25519 private key bytes and tags with the ed25519
// signer multiformat code, returning an ed25519 signer.
func FromRaw(b []byte) (Ed25519Signer, error) {
	if len(b) != ed25519.SeedSize {
		return nil, fmt.Errorf("invalid length: %d wanted: %d", len(b), ed25519.SeedSize)
	}
	s := make(Ed25519Signer, size)
	varint.PutUvarint(s, Code)
	copy(s[tagSize:size], b[:ed25519.SeedSize])
	return s, nil
}

type Ed25519Signer []byte

var _ principal.Signer = (Ed25519Signer)(nil)

func (s Ed25519Signer) Code() uint64 {
	return Code
}

func (s Ed25519Signer) SignatureCode() uint64 {
	return SignatureCode
}

func (s Ed25519Signer) Verifier() principal.Verifier {
	sk := ed25519.NewKeyFromSeed(s[tagSize:])
	v, _ := verifier.FromRaw(sk.Public().(ed25519.PublicKey))
	return v
}

func (s Ed25519Signer) DID() did.DID {
	return s.Verifier().DID()
}

// Bytes returns the private key bytes with multiformat prefix varint.
func (s Ed25519Signer) Bytes() []byte {
	return s
}

// Raw encodes the bytes of the private key without multiformats tags.
func (s Ed25519Signer) Raw() []byte {
	pk := make([]byte, keySize)
	copy(pk, s[tagSize:size])
	return pk
}

func (s Ed25519Signer) Sign(msg []byte) []byte {
	sk := ed25519.NewKeyFromSeed(s[tagSize:])
	return ed25519.Sign(sk, msg)
}
