package ucan

import (
	"crypto"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan/crypto/signature"
)

// The Principal who's authority is delegated or invoked. A Subject represents
// the Agent that a capability is for. A Subject MUST be referenced by DID.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#subject
type Subject = did.DID

// Commands are concrete messages ("verbs") that MUST be unambiguously
// interpretable by the Subject of a UCAN.
//
// Commands MUST be lowercase, and begin with a slash (/). Segments MUST be
// separated by a slash. A trailing slash MUST NOT be present.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#command
type Command = string

// Capability represents an ability that a UCAN holder can perform with some
// resource.
type Capability interface {
	Subject() Subject
	Command() Command
	Policy() []string // TODO: define properly
}

// Principal is a DID object representation with a `did` accessor for the DID.
type Principal interface {
	DID() did.DID
}

// UTCUnixTimestamp is a timestamp in seconds since the Unix epoch.
type UTCUnixTimestamp = uint64

// https://github.com/ucan-wg/spec/blob/main/README.md#nonce
type Nonce = []byte

// Signer is an entity that can sign UCANs with keys from a `Principal`.
type Signer interface {
	Principal
	crypto.Signer

	// SignatureCode is an integer corresponding to the byteprefix of the
	// signature algorithm. It is used by varsig to tag the signature so it can
	// self describe what algorithm was used.
	SignatureCode() uint64
}

// Verifier is an entity that can verify UCAN signatures against a `Principal`.
type Verifier interface {
	Principal
	signature.Verifier
}
