package eddsa

import (
	"fmt"

	varint "github.com/multiformats/go-varint"
)

const Code = 0xED

type SignatureAlgorithm struct {
	curve    uint64
	hashAlgo uint64
}

func New(curve uint64, hashAlgo uint64) SignatureAlgorithm {
	return SignatureAlgorithm{curve, hashAlgo}
}

func (sa SignatureAlgorithm) Code() uint64 {
	return Code
}

func (sa SignatureAlgorithm) Curve() uint64 {
	return sa.curve
}

func (sa SignatureAlgorithm) HashAlgorithm() uint64 {
	return sa.hashAlgo
}

type Codec struct {
	curve    uint64
	hashAlgo uint64
}

func NewCodec(curve uint64, hashAlgo uint64) Codec {
	return Codec{curve, hashAlgo}
}

func (sac Codec) Code() uint64 {
	return Code
}

func (sac Codec) Curve() uint64 {
	return sac.curve
}

func (sac Codec) HashAlgorithm() uint64 {
	return sac.hashAlgo
}

func (sac Codec) Encode(enc SignatureAlgorithm) ([]byte, error) {
	size := varint.UvarintSize(Code)
	size += varint.UvarintSize(enc.curve)
	size += varint.UvarintSize(enc.hashAlgo)
	out := make([]byte, size)
	offset := varint.PutUvarint(out, Code)
	offset += varint.PutUvarint(out[offset:], enc.curve)
	varint.PutUvarint(out[offset:], enc.hashAlgo)
	return out, nil
}

func (sac Codec) Decode(input []byte) (SignatureAlgorithm, int, error) {
	code, n, err := varint.FromUvarint(input)
	if err != nil {
		return SignatureAlgorithm{}, 0, err
	}
	if code != Code {
		return SignatureAlgorithm{}, n, fmt.Errorf("signature code is not EdDSA: 0x%02x, expected: 0x%02x", code, Code)
	}
	offset := n

	curve, n, err := varint.FromUvarint(input[offset:])
	if err != nil {
		return SignatureAlgorithm{}, 0, err
	}
	if curve != sac.curve {
		return SignatureAlgorithm{}, n, fmt.Errorf("unexpected curve code: 0x%02x, expected: 0x%02x", curve, sac.curve)
	}
	offset += n

	hashAlgo, n, err := varint.FromUvarint(input[offset:])
	if err != nil {
		return SignatureAlgorithm{}, 0, err
	}
	if curve != sac.curve {
		return SignatureAlgorithm{}, n, fmt.Errorf("unexpected hash algorithm code: 0x%02x, expected: 0x%02x", hashAlgo, sac.hashAlgo)
	}
	offset += n

	return SignatureAlgorithm{curve, hashAlgo}, offset, nil
}
