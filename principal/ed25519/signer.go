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

var privateTagSize = varint.UvarintSize(Code)
var publicTagSize = varint.UvarintSize(verifier.Code)

const keySize = 32

var size = privateTagSize + keySize + publicTagSize + keySize
var pubKeyOffset = privateTagSize + keySize

func Generate() (Ed25519Signer, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating Ed25519 key: %w", err)
	}
	s := make(Ed25519Signer, size)
	varint.PutUvarint(s, Code)
	copy(s[privateTagSize:], priv)
	varint.PutUvarint(s[pubKeyOffset:], verifier.Code)
	copy(s[pubKeyOffset+publicTagSize:], pub)
	return s, nil
}

func Parse(str string) (Ed25519Signer, error) {
	_, bytes, err := multibase.Decode(str)
	if err != nil {
		return nil, fmt.Errorf("decoding multibase string: %w", err)
	}
	return Decode(bytes)
}

func Format(signer principal.Signer) (string, error) {
	return multibase.Encode(multibase.Base64pad, signer.Bytes())
}

func Decode(b []byte) (Ed25519Signer, error) {
	if len(b) != size {
		return nil, fmt.Errorf("invalid length: %d wanted: %d", len(b), size)
	}

	prc, _, err := varint.FromUvarint(b)
	if err != nil {
		return nil, fmt.Errorf("reading private key uvarint: %w", err)
	}
	if prc != Code {
		return nil, fmt.Errorf("invalid private key codec: 0x%02x, expected: 0x%02x", prc, Code)
	}

	puc, _, err := varint.FromUvarint(b[pubKeyOffset:])
	if err != nil {
		return nil, fmt.Errorf("reading public key uvarint: %w", err)
	}
	if puc != verifier.Code {
		return nil, fmt.Errorf("invalid public key codec: 0x%02x, expected: 0x%02x", puc, Code)
	}

	_, err = verifier.Decode(b[pubKeyOffset:])
	if err != nil {
		return nil, fmt.Errorf("decoding public key: %w", err)
	}

	s := make(Ed25519Signer, size)
	copy(s, b)

	return s, nil
}

// FromRaw takes raw ed25519 private key bytes and tags with the ed25519 signer
// and verifier multiformat codes, returning an ed25519 signer.
func FromRaw(b []byte) (principal.Signer, error) {
	if len(b) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid length: %d wanted: %d", len(b), ed25519.PrivateKeySize)
	}
	s := make(Ed25519Signer, size)
	varint.PutUvarint(s, Code)
	copy(s[privateTagSize:privateTagSize+keySize], b[:ed25519.PrivateKeySize-ed25519.PublicKeySize])
	varint.PutUvarint(s[pubKeyOffset:], verifier.Code)
	copy(s[pubKeyOffset+publicTagSize:], b[ed25519.PrivateKeySize-ed25519.PublicKeySize:ed25519.PrivateKeySize])
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
	return verifier.Ed25519Verifier(s[pubKeyOffset:])
}

func (s Ed25519Signer) DID() did.DID {
	id, _ := did.Decode(s[pubKeyOffset:])
	return id
}

// Bytes returns the private key bytes with multiformat prefix varint.
func (s Ed25519Signer) Bytes() []byte {
	return s
}

// Raw encodes the bytes of the public key without multiformats tags.
func (s Ed25519Signer) Raw() []byte {
	pk := make(ed25519.PrivateKey, ed25519.PrivateKeySize)
	copy(pk[0:ed25519.PublicKeySize], s[privateTagSize:pubKeyOffset])
	copy(pk[ed25519.PrivateKeySize-ed25519.PublicKeySize:ed25519.PrivateKeySize], s[pubKeyOffset+publicTagSize:pubKeyOffset+publicTagSize+keySize])
	return pk
}

func (s Ed25519Signer) Sign(msg []byte) []byte {
	return ed25519.Sign(s.Raw(), msg)
}
