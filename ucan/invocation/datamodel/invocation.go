package datamodel

import (
	"io"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan"
	edm "github.com/alanshaw/ucantone/ucan/envelope/datamodel"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

// CBORMarshalable is an interface describing a type that allows both
// marshalling to CBOR as well as unmarshalling from CBOR.
type CBORMarshalable interface {
	cbg.CBORMarshaler
	cbg.CBORUnmarshaler
}

type StructModel struct {
	Value CBORMarshalable
}

func (sm *StructModel) MarshalCBOR(w io.Writer) error {
	return sm.Value.MarshalCBOR(w)
}

func (sm *StructModel) UnmarshalCBOR(r io.Reader) error {
	return sm.Value.UnmarshalCBOR(r)
}

var _ CBORMarshalable = (*StructModel)(nil)

type ArgsModel = StructModel

type MetaModel = StructModel

type TokenPayloadModel1_0_0_rc1 struct {
	// Issuer DID (sender).
	Iss did.DID `cborgen:"iss"`
	// The Subject being invoked.
	Sub ucan.Subject `cborgen:"sub"`
	// The DID of the intended Executor if different from the Subject.
	Aud *did.DID `cborgen:"aud"`
	// The command to invoke.
	Cmd ucan.Command `cborgen:"cmd"`
	// The command arguments.
	Args *ArgsModel `cborgen:"args"`
	// Delegations that prove the chain of authority
	Prf []cid.Cid `cborgen:"prf"`
	// Arbitrary metadata.
	Meta *MetaModel `cborgen:"meta"`
	// A unique, random nonce.
	Nonce ucan.Nonce `cborgen:"nonce"`
	// The timestamp at which the Invocation becomes invalid.
	Exp *ucan.UTCUnixTimestamp `cborgen:"exp"`
	// The timestamp at which the Invocation was created
	Iat *ucan.UTCUnixTimestamp `cborgen:"iat"`
	// An OPTIONAL CID of the Receipt that enqueued the Task.
	Cause *cid.Cid `cborgen:"cause"`
}

type SigPayloadModel struct {
	// The Varsig v1 header.
	Header []byte `cborgen:"h"`
	// The UCAN token payload.
	TokenPayload1_0_0_rc1 *TokenPayloadModel1_0_0_rc1 `cborgen:"ucan/inv@1.0.0-rc.1,omitempty"`
}

type EnvelopeModel edm.EnvelopeModel[SigPayloadModel]
