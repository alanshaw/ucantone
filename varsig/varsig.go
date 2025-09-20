package varsig

import (
	"fmt"

	varint "github.com/multiformats/go-varint"
)

const Prefix = 0x34
const Version = 0x01

type Codec[T any] interface {
	Code() uint64
	Encode(T) ([]byte, error)
	Decode([]byte) (T, int, error)
}

type SignatureAlgorithm interface {
	// Discriminant for the signature segments.
	Code() uint64
}

type SignatureAlgorithmCodec[T SignatureAlgorithm] Codec[T]

var signatureAlgorithmCodecs = map[uint64]SignatureAlgorithmCodec[SignatureAlgorithm]{}

type signatureAlgorithmCodecAdapter[T SignatureAlgorithm] struct {
	codec SignatureAlgorithmCodec[T]
}

func (a signatureAlgorithmCodecAdapter[T]) Code() uint64 {
	return a.codec.Code()
}

func (a signatureAlgorithmCodecAdapter[T]) Encode(algo SignatureAlgorithm) ([]byte, error) {
	return a.codec.Encode(algo.(T))
}

func (a signatureAlgorithmCodecAdapter[T]) Decode(input []byte) (SignatureAlgorithm, int, error) {
	algo, n, err := a.codec.Decode(input)
	if err != nil {
		return nil, 0, err
	}
	return SignatureAlgorithm(algo), n, nil
}

func RegisterSignatureAlgorithm[T SignatureAlgorithm](codec SignatureAlgorithmCodec[T]) {
	signatureAlgorithmCodecs[codec.Code()] = signatureAlgorithmCodecAdapter[T]{codec}
}

type PayloadEncoding interface {
	// Discriminant for the payload encoding segments.
	Code() uint64
}

type PayloadEncodingCodec[T PayloadEncoding] Codec[T]

var payloadEncodingCodecs = map[uint64]PayloadEncodingCodec[PayloadEncoding]{}

type payloadEncodingCodecAdapter[T PayloadEncoding] struct {
	codec PayloadEncodingCodec[T]
}

func (a payloadEncodingCodecAdapter[T]) Code() uint64 {
	return a.codec.Code()
}

func (a payloadEncodingCodecAdapter[T]) Encode(algo PayloadEncoding) ([]byte, error) {
	return a.codec.Encode(algo.(T))
}

func (a payloadEncodingCodecAdapter[T]) Decode(input []byte) (PayloadEncoding, int, error) {
	algo, n, err := a.codec.Decode(input)
	if err != nil {
		return nil, 0, err
	}
	return PayloadEncoding(algo), n, nil
}

func RegisterPayloadEncoding[T SignatureAlgorithm](codec PayloadEncodingCodec[T]) {
	payloadEncodingCodecs[codec.Code()] = payloadEncodingCodecAdapter[T]{codec}
}

type VarsigHeader[S SignatureAlgorithm, P PayloadEncoding] interface {
	// A Varsig v1 MUST use the 0x01 version tag.
	Version() uint64
	SignatureAlgorithm() S
	PayloadEncoding() P
}

type Header[S SignatureAlgorithm, P PayloadEncoding] struct {
	signatureAlgorithm S
	payloadEncoding    P
}

func NewHeader[S SignatureAlgorithm, P PayloadEncoding](sigAlgo S, payloadEnc P) Header[S, P] {
	return Header[S, P]{sigAlgo, payloadEnc}
}

func (h Header[S, P]) Version() uint64 {
	return Version
}

func (h Header[S, P]) SignatureAlgorithm() S {
	return h.signatureAlgorithm
}

func (h Header[S, P]) PayloadEncoding() P {
	return h.payloadEncoding
}

var _ VarsigHeader[SignatureAlgorithm, PayloadEncoding] = (*Header[SignatureAlgorithm, PayloadEncoding])(nil)

func Encode[S SignatureAlgorithm, P PayloadEncoding](header VarsigHeader[S, P]) ([]byte, error) {
	size := varint.UvarintSize(Prefix)
	size += varint.UvarintSize(Version)

	sigAlgoCodec, ok := signatureAlgorithmCodecs[header.SignatureAlgorithm().Code()]
	if !ok {
		return nil, fmt.Errorf("missing codec for signature algorithm: %d", header.SignatureAlgorithm().Code())
	}
	sigAlgoBytes, err := sigAlgoCodec.Encode(header.SignatureAlgorithm())
	if err != nil {
		return nil, err
	}
	size += len(sigAlgoBytes)

	payloadEncCodec, ok := payloadEncodingCodecs[header.PayloadEncoding().Code()]
	if !ok {
		return nil, fmt.Errorf("missing codec for payload encoding: %d", header.PayloadEncoding().Code())
	}
	payloadEncBytes, err := payloadEncCodec.Encode(header.PayloadEncoding())
	if err != nil {
		return nil, err
	}
	size += len(payloadEncBytes)

	out := make([]byte, size)
	offset := varint.PutUvarint(out, Prefix)
	offset += varint.PutUvarint(out[offset:], Version)
	offset += copy(out[offset:], sigAlgoBytes)
	offset += copy(out[offset:], payloadEncBytes)
	return out, nil
}

func Decode(input []byte) (Header[SignatureAlgorithm, PayloadEncoding], error) {
	offset := 0
	prefix, n, err := varint.FromUvarint(input)
	if err != nil {
		return Header[SignatureAlgorithm, PayloadEncoding]{}, fmt.Errorf("reading prefix: %w", err)
	}
	if prefix != Prefix {
		return Header[SignatureAlgorithm, PayloadEncoding]{}, fmt.Errorf("invalid varsig prefix: 0x%02x, expected: 0x%02x", prefix, Prefix)
	}
	offset += n

	version, n, err := varint.FromUvarint(input[offset:])
	if err != nil {
		return Header[SignatureAlgorithm, PayloadEncoding]{}, fmt.Errorf("reading version: %w", err)
	}
	if version != Version {
		return Header[SignatureAlgorithm, PayloadEncoding]{}, fmt.Errorf("invalid varsig version: 0x%02x, expected: 0x%02x", version, Version)
	}
	offset += n

	sigAlgoCode, _, err := varint.FromUvarint(input[offset:])
	if err != nil {
		return Header[SignatureAlgorithm, PayloadEncoding]{}, fmt.Errorf("reading signature algorithm code: %w", err)
	}
	sigAlgoCodec, ok := signatureAlgorithmCodecs[sigAlgoCode]
	if !ok {
		return Header[SignatureAlgorithm, PayloadEncoding]{}, fmt.Errorf("missing codec for signature algorithm: 0x%02x", sigAlgoCode)
	}
	sigAlgo, n, err := sigAlgoCodec.Decode(input[offset:])
	if err != nil {
		return Header[SignatureAlgorithm, PayloadEncoding]{}, fmt.Errorf("decoding signature algorithm: %w", err)
	}
	offset += n

	payloadEncCode, _, err := varint.FromUvarint(input[offset:])
	if err != nil {
		return Header[SignatureAlgorithm, PayloadEncoding]{}, fmt.Errorf("reading payload encoding code: %w", err)
	}
	payloadEncCodec, ok := payloadEncodingCodecs[payloadEncCode]
	if !ok {
		return Header[SignatureAlgorithm, PayloadEncoding]{}, fmt.Errorf("missing codec for payload encoding: 0x%02x", sigAlgoCode)
	}
	payloadEnc, _, err := payloadEncCodec.Decode(input[offset:])
	if err != nil {
		return Header[SignatureAlgorithm, PayloadEncoding]{}, fmt.Errorf("decoding payload encoding: %w", err)
	}

	return Header[SignatureAlgorithm, PayloadEncoding]{sigAlgo, payloadEnc}, nil
}
