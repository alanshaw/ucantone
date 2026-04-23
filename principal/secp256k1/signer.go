package secp256k1

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/principal"
	"github.com/alanshaw/ucantone/principal/secp256k1/verifier"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-varint"
)

const Code = 0x1301

var SignatureAlgorithm = verifier.SignatureAlgorithm

var tagSize = varint.UvarintSize(Code)

const keySize = 32

var size = tagSize + keySize

func Generate() (Signer, error) {
	priv, err := ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating secp256k1 key: %w", err)
	}
	s := make(Signer, size)
	varint.PutUvarint(s, Code)
	priv.D.FillBytes(s[tagSize:])
	return s, nil
}

// Parse parses a multibase encoded string containing a secp256k1 signer
// multiformat varint (0x1301) +  byte secp256k1 raw scalar value.
func Parse(str string) (Signer, error) {
	_, bytes, err := multibase.Decode(str)
	if err != nil {
		return nil, fmt.Errorf("decoding multibase string: %w", err)
	}
	return Decode(bytes)
}

// Decode decodes a buffer of a secp256k1 signer multiformat varint (0x1301) +
// 32 byte secp256k1 raw scalar value.
func Decode(b []byte) (Signer, error) {
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
	s := make(Signer, size)
	copy(s, b)
	return s, nil
}

// FromRaw takes raw 32 byte scalar value and tags with the secp256k1
// signer multiformat code, returning a secp256k1 signer.
func FromRaw(b []byte) (Signer, error) {
	if len(b) != keySize {
		return nil, fmt.Errorf("invalid length: %d wanted: %d", len(b), keySize)
	}
	s := make(Signer, size)
	varint.PutUvarint(s, Code)
	copy(s[tagSize:], b)
	return s, nil
}

type Signer []byte

var _ principal.Signer = (Signer)(nil)

func (s Signer) Code() uint64 {
	return Code
}

func (s Signer) SignatureAlgorithm() varsig.SignatureAlgorithm {
	return SignatureAlgorithm
}

func (s Signer) Verifier() principal.Verifier {
	x, y := secp256k1.S256().ScalarBaseMult(s[tagSize:])
	v, _ := verifier.FromRaw(secp256k1.CompressPubkey(x, y))
	return v
}

func (s Signer) DID() did.DID {
	return s.Verifier().DID()
}

// Bytes returns the private key bytes with multiformat prefix varint.
func (s Signer) Bytes() []byte {
	return s
}

// Raw encodes the bytes of the private key without multiformats tags.
func (s Signer) Raw() []byte {
	pk := make([]byte, keySize)
	copy(pk, s[tagSize:size])
	return pk
}

func (s Signer) Sign(msg []byte) []byte {
	hash := sha256.New()
	hash.Write(msg)
	sig, _ := secp256k1.Sign(hash.Sum(nil), s[tagSize:])
	return sig[:crypto.RecoveryIDOffset]
}
