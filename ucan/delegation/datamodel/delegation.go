package datamodel

import (
	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan"
	edm "github.com/alanshaw/ucantone/ucan/envelope/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

const Tag = "ucan/dlg@1.0.0-rc.1"

type TokenPayloadModel1_0_0_rc1 struct {
	// Issuer DID (sender).
	Iss did.DID `cborgen:"iss"`
	// The DID of the intended Executor if different from the Subject.
	Aud did.DID `cborgen:"aud"`
	// The principal the chain is about.
	Sub did.DID `cborgen:"sub"`
	// The command to eventually invoke.
	Cmd ucan.Command `cborgen:"cmd"`
	// TODO
	Pol []string `cborgen:"pol"`
	// A unique, random nonce.
	Nonce ucan.Nonce `cborgen:"nonce"`
	// Arbitrary metadata.
	Meta *cbg.Deferred `cborgen:"meta,omitempty"`
	// "Not before" UTC Unix Timestamp in seconds (valid from).
	Nbf *ucan.UTCUnixTimestamp `cborgen:"nbf,omitempty"`
	// Expiration UTC Unix Timestamp in seconds (valid until).
	Exp *ucan.UTCUnixTimestamp `cborgen:"exp"`
}

type SigPayloadModel struct {
	// The Varsig v1 header.
	Header []byte `cborgen:"h"`
	// The UCAN token payload.
	TokenPayload1_0_0_rc1 *TokenPayloadModel1_0_0_rc1 `cborgen:"ucan/dlg@1.0.0-rc.1,omitempty"`
}

type EnvelopeModel edm.EnvelopeModel[SigPayloadModel]
