package ucan

import (
	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan/crypto"
	"github.com/alanshaw/ucantone/ucan/crypto/signature"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/ipfs/go-cid"
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

type Signature interface {
	Header() varsig.VarsigHeader[varsig.SignatureAlgorithm, varsig.PayloadEncoding]
	Bytes() []byte
}

// Verifier is an entity that can verify UCAN signatures against a `Principal`.
type Verifier interface {
	Principal
	signature.Verifier
}

// Link is an IPLD link to a UCAN token.
type Link = cid.Cid

type UCAN interface {
	// Issuer DID (sender).
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
	Issuer() Principal
	// The Subject being invoked.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#subject
	Subject() Principal
	// The DID of the intended Executor if different from the Subject.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
	Audience() Principal
	// The command to invoke.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#command
	Command() Command
	// The command arguments.
	Args() any
	// Delegations that prove the chain of authority.
	Proofs() []Link
	// Arbitrary metadata.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#metadata
	Metadata() any
	// A unique, random nonce.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#nonce
	Nonce() Nonce
	// The timestamp at which the invocation becomes invalid.
	Expiration() *UTCUnixTimestamp
	// Signature of the UCAN issuer.
	Signature() Signature
}

// A capability is the semantically-relevant claim of a delegation.
//
// https://github.com/ucan-wg/delegation/blob/main/README.md#capability
type Capability interface {
	// The Subject that this capability is about.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#subject
	Subject() Principal
	// The command of this capability.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#command
	Command() Command
	// Additional constraints on eventual invocation arguments, expressed in the
	// UCAN Policy Language.
	//
	// https://github.com/ucan-wg/delegation/blob/main/README.md#policy
	Policy() []string // TODO define properly
}

// UCAN Delegation is a delegable certificate capability system with
// runtime-extensibility, ad hoc conditions, cacheability, and focused on ease
// of use and interoperability. Delegations act as a proofs for UCAN
// Invocations.
//
// https://github.com/ucan-wg/delegation/blob/main/README.md
type Delegation interface {
	UCAN
	Capability
}

// UCAN Invocation defines a format for expressing the intention to execute
// delegated UCAN capabilities, and the attested receipts from an execution.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md
type Invocation interface {
	UCAN
	// The timestamp at which the invocation was created.
	IssuedAt() *UTCUnixTimestamp
	// CID of the receipt that enqueued the Task.
	Cause() *Link
}

type Receipt interface {
	Invocation
}

// Container is a format for transmitting one or more UCAN tokens as bytes,
// regardless of the transport.
//
// https://github.com/ucan-wg/container/blob/main/Readme.md
type Container interface {
	Invocations() []Invocation
	Delegations() []Delegation
	Delegation(Link) (Delegation, error)
	Receipts() []Receipt
	Receipt(Link) (Receipt, error)
}
