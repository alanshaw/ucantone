package ucan

import (
	"time"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/crypto"
	"github.com/alanshaw/ucantone/ucan/crypto/signature"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/ipfs/go-cid"
)

// The Principal who's authority is delegated or invoked. A Subject represents
// the Agent that a capability is for. A Subject MUST be referenced by DID.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#subject
type Subject = Principal

// Commands are concrete messages ("verbs") that MUST be unambiguously
// interpretable by the Subject of a UCAN.
//
// Commands MUST be lowercase, and begin with a slash (/). Segments MUST be
// separated by a slash. A trailing slash MUST NOT be present.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#command
type Command = command.Command

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

// Signature encapsulates the bytes that comprise the signature as well as the
// details of the signing algorithm and payload encoding.
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
	ipld.Block
	// Issuer DID (sender).
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
	Issuer() Principal
	// The subject being invoked.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#subject
	Subject() Principal
	// Audience can be conceptualized as the receiver of a postal letter.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
	Audience() Principal
	// The command to eventually invoke.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#command
	Command() Command
	// Arbitrary metadata.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#metadata
	Metadata() ipld.Map[string, ipld.Any]
	// A unique, random nonce.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#nonce
	Nonce() Nonce
	// The timestamp at which the invocation becomes invalid.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#time-bounds
	Expiration() *UTCUnixTimestamp
	// Signature of the UCAN issuer.
	Signature() Signature
}

// UCAN Delegation uses predicate logic statements extended with jq-inspired
// selectors as a policy language. Policies are syntactically driven, and
// constrain the args field of an eventual Invocation.
//
// https://github.com/ucan-wg/delegation/blob/main/README.md#policy
type Policy = policy.Policy

// A capability is the semantically-relevant claim of a delegation.
//
// https://github.com/ucan-wg/delegation/blob/main/README.md#capability
type Capability interface {
	// The subject that this capability is about.
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
	Policy() Policy
}

// UCAN Delegation is a delegable certificate capability system with
// runtime-extensibility, ad-hoc conditions, cacheability, and focused on ease
// of use and interoperability. Delegations act as a proofs for UCAN
// invocations.
//
// https://github.com/ucan-wg/delegation/blob/main/README.md
type Delegation interface {
	Capability
	UCAN
	// NotBefore is the time in seconds since the Unix epoch that the UCAN
	// becomes valid.
	//
	// https://github.com/ucan-wg/spec/blob/main/README.md#time-bounds
	NotBefore() *UTCUnixTimestamp
}

// A Task is the subset of Invocation fields that uniquely determine the work to
// be performed.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#task
type Task interface {
	// A concrete, dispatchable message that can be sent to the Executor.
	//
	// https://github.com/ucan-wg/invocation/blob/main/README.md#command
	Command() Command
	// The subject being invoked.
	//
	// https://github.com/ucan-wg/invocation/blob/main/README.md#subject
	Subject() Principal
	// Parameters expected by the command.
	//
	// https://github.com/ucan-wg/invocation/blob/main/README.md#arguments
	Arguments() ipld.Map[string, ipld.Any]
	// A unique, random nonce. It ensures that multiple (non-idempotent)
	// invocations are unique. The nonce SHOULD be empty (0x) for commands that
	// are idempotent (such as deterministic Wasm modules or standards-abiding
	// HTTP PUT requests).
	//
	// https://github.com/ucan-wg/invocation/blob/main/README.md#nonce
	Nonce() Nonce
}

// UCAN Invocation defines a format for expressing the intention to execute
// delegated UCAN capabilities, and the attested receipts from an execution.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md
type Invocation interface {
	Task
	UCAN
	// Task returns the CID of the fields that comprise the task for the
	// invocation.
	//
	// https://github.com/ucan-wg/invocation/blob/main/README.md#task
	Task() cid.Cid
	// Delegations that prove the chain of authority.
	//
	// https://github.com/ucan-wg/invocation/blob/main/README.md#proofs
	Proofs() []Link
	// The timestamp at which the invocation was created.
	//
	// https://github.com/ucan-wg/invocation/blob/main/README.md#issued-at
	IssuedAt() *UTCUnixTimestamp
	// CID of the receipt that enqueued the Task.
	//
	// https://github.com/ucan-wg/invocation/blob/main/README.md#cause
	Cause() *Link
}

// UCAN Invocation Receipt is a signed assertion of the executor state
// describing the result and effects of the invocation.
type Receipt interface {
	UCAN
	Invocation
	// Ran is the CID of the executed task the receipt is for.
	Ran() cid.Cid
	// Out is the attested result of the execution of the task.
	Out() result.Result[ipld.Any, ipld.Any]
}

// Container is a format for transmitting one or more UCAN tokens as bytes,
// regardless of the transport.
//
// https://github.com/ucan-wg/container/blob/main/Readme.md
type Container interface {
	// Invocations the container contains.
	Invocations() []Invocation
	// Delegations the container contains.
	Delegations() []Delegation
	// Delegation retrieves a delegation from the container by it's CID.
	Delegation(Link) (Delegation, error)
	// Receipts the container contains.
	Receipts() []Receipt
	// Receipt retrieves a receipt from the container by the CID of a [Task] that
	// was executed.
	Receipt(Link) (Receipt, error)
}

// IsExpired checks if a UCAN is expired.
func IsExpired(ucan UCAN) bool {
	exp := ucan.Expiration()
	if exp == nil {
		return false
	}
	return *exp <= Now()
}

// IsTooEarly checks if a delegation is not active yet.
func IsTooEarly(delegation Delegation) bool {
	nbf := delegation.NotBefore()
	if nbf == nil {
		return false
	}
	return *nbf != 0 && Now() <= *nbf
}

// Now returns a UTC Unix timestamp for comparing it against time window of the
// UCAN.
func Now() UTCUnixTimestamp {
	return UTCUnixTimestamp(time.Now().Unix())
}
