package datamodel

import (
	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan"
	edm "github.com/alanshaw/ucantone/ucan/envelope/datamodel"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

const Tag = "ucan/inv@1.0.0-rc.1"

type TokenPayloadModel1_0_0_rc1 struct {
	// Issuer DID (sender).
	Iss did.DID `cborgen:"iss"`
	// The Subject being invoked.
	Sub did.DID `cborgen:"sub"`
	// The DID of the intended Executor if different from the Subject.
	Aud *did.DID `cborgen:"aud,omitempty"`
	// The command to invoke.
	Cmd ucan.Command `cborgen:"cmd"`
	// The command arguments.
	Args cbg.Deferred `cborgen:"args"`
	// Delegations that prove the chain of authority.
	Prf []cid.Cid `cborgen:"prf"`
	// Arbitrary metadata.
	Meta *cbg.Deferred `cborgen:"meta,omitempty"`
	// A unique, random nonce.
	Nonce ucan.Nonce `cborgen:"nonce"`
	// The timestamp at which the invocation becomes invalid.
	Exp *ucan.UTCUnixTimestamp `cborgen:"exp"`
	// The timestamp at which the invocation was created.
	Iat *ucan.UTCUnixTimestamp `cborgen:"iat,omitempty"`
	// CID of the receipt that enqueued the Task.
	Cause *cid.Cid `cborgen:"cause,omitempty"`
}

type SigPayloadModel struct {
	// The Varsig v1 header.
	Header []byte `cborgen:"h"`
	// The UCAN token payload.
	TokenPayload1_0_0_rc1 *TokenPayloadModel1_0_0_rc1 `cborgen:"ucan/inv@1.0.0-rc.1,omitempty"`
}

type EnvelopeModel edm.EnvelopeModel[SigPayloadModel]
