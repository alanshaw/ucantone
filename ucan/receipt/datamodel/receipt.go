package datamodel

import (
	"github.com/alanshaw/ucantone/did"
	rdm "github.com/alanshaw/ucantone/result/datamodel"
	"github.com/alanshaw/ucantone/ucan"
	edm "github.com/alanshaw/ucantone/ucan/envelope/datamodel"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

const Tag = "ucan/rcpt@1.0.0-rc.1"

type ArgsModel struct {
	// Ran is the CID of the executed task the receipt is for.
	Ran cid.Cid `cborgen:"ran"`
	// Out is the attested result of the execution of the task.
	Out rdm.ResultModel `cborgen:"out"`
	// TODO: add Run
}

type TokenPayloadModel1_0_0_rc1 struct {
	// DID of the Invocation Executor.
	Iss did.DID `cborgen:"iss"`
	// DID of the Invocation Executor.
	Sub did.DID `cborgen:"sub"`
	// DID of the Invocation Executor.
	Aud did.DID `cborgen:"aud"`
	// Constant "/ucan/assert/receipt"
	Cmd ucan.Command `cborgen:"cmd,const=/ucan/assert/receipt"`
	// The command arguments.
	Args ArgsModel `cborgen:"args"`
	// Delegations that prove the chain of authority.
	Prf []cid.Cid `cborgen:"prf"`
	// Arbitrary metadata.
	Meta *cbg.Deferred `cborgen:"meta,omitempty"`
	// A unique, random nonce.
	Nonce ucan.Nonce `cborgen:"nonce"`
	// Denotes the time until which Executor is commited to uphold issued assertion.
	Exp *ucan.UTCUnixTimestamp `cborgen:"exp"`
	// The timestamp at which the receipt was created.
	Iat *ucan.UTCUnixTimestamp `cborgen:"iat,omitempty"`
}

type SigPayloadModel struct {
	// The Varsig v1 header.
	Header []byte `cborgen:"h"`
	// The UCAN token payload.
	TokenPayload1_0_0_rc1 *TokenPayloadModel1_0_0_rc1 `cborgen:"ucan/rcpt@1.0.0-rc.1,omitempty"`
}

type EnvelopeModel edm.EnvelopeModel[SigPayloadModel]
