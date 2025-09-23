package common

import (
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/algorithm/ed25519"
	"github.com/alanshaw/ucantone/varsig/payload/dagcbor"
)

func init() {
	varsig.RegisterSignatureAlgorithm(ed25519.NewCodec())
	varsig.RegisterPayloadEncoding(dagcbor.NewCodec())
}

var DagCbor = dagcbor.New()
var Ed25519 = ed25519.New()

var Ed25519DagCbor = varsig.NewHeader(Ed25519, DagCbor)
