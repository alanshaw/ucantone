package ed25519

import (
	"github.com/alanshaw/ucantone/varsig/algorithm/eddsa"
)

const Code = 0xED
const Sha2_512 = 0x13

type SignatureAlgorithm = eddsa.SignatureAlgorithm

func New() SignatureAlgorithm {
	return eddsa.New(Code, Sha2_512)
}

type Codec = eddsa.Codec

func NewCodec() Codec {
	return eddsa.NewCodec(Code, Sha2_512)
}
