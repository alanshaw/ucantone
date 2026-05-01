package absentee

import (
	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/algorithm/nonstandard"
)

var SignatureAlgorithm = nonstandard.New()

// Signer is a special type of signer that produces an absent signature,
// which signals that verifier needs to verify authorization interactively.
type Signer struct {
	id did.DID
}

var _ ucan.Signer = Signer{}

func (a Signer) DID() did.DID {
	return a.id
}

func (a Signer) Sign(msg []byte) []byte {
	return []byte{}
}

func (a Signer) SignatureAlgorithm() varsig.SignatureAlgorithm {
	return SignatureAlgorithm
}

// From creates a special type of signer that produces an absent signature,
// which signals that verifier needs to verify authorization interactively.
func From(id did.DID) Signer {
	return Signer{id}
}
