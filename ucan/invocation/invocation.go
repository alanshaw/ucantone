package invocation

import (
	"bytes"
	"fmt"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/crypto/signature"
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/algorithm/ed25519"
	"github.com/alanshaw/ucantone/varsig/common"
	cid "github.com/ipfs/go-cid"
)

type Invocation struct {
	sig   *signature.Signature
	model *idm.EnvelopeModel
}

// Parameters expected by the command.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#arguments
func (inv *Invocation) Arguments() any {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Args
}

// The DID of the intended Executor if different from the Subject.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
func (inv *Invocation) Audience() ucan.Principal {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Aud
}

// A provenance claim describing which receipt requested it.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#cause
func (inv *Invocation) Cause() *cid.Cid {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Cause
}

// The command to invoke.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#command
func (inv *Invocation) Command() ucan.Command {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Cmd
}

// The timestamp at which the invocation becomes invalid.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#expiration
func (inv *Invocation) Expiration() *ucan.UTCUnixTimestamp {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Exp
}

// An issuance timestamp.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#issued-at
func (inv *Invocation) IssuedAt() *ucan.UTCUnixTimestamp {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Iat
}

// Issuer DID (sender).
//
// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
func (inv *Invocation) Issuer() ucan.Principal {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Iss
}

// Arbitrary metadata or extensible fields.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#metadata
func (inv *Invocation) Metadata() any {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Meta
}

// The datamodel this invocation is built from.
func (inv *Invocation) Model() *idm.EnvelopeModel {
	return inv.model
}

// A unique, random nonce. It ensures that multiple (non-idempotent) invocations
// are unique. The nonce SHOULD be empty (0x) for Commands that are idempotent
// (such as deterministic Wasm modules or standards-abiding HTTP PUT requests).
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#nonce
func (inv *Invocation) Nonce() ucan.Nonce {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Nonce
}

// The path of authority from the subject to the invoker.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#proofs
func (inv *Invocation) Proofs() []cid.Cid {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Prf
}

// The signature over the payload.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#envelope
func (inv *Invocation) Signature() ucan.Signature {
	return inv.sig
}

// The Subject being invoked.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#subject
func (inv *Invocation) Subject() ucan.Principal {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Sub
}

var _ ucan.Invocation = (*Invocation)(nil)

func Encode(inv ucan.Invocation) ([]byte, error) {
	sigHeader, err := varsig.Encode(inv.Signature().Header())
	if err != nil {
		return nil, fmt.Errorf("encoding varsig header: %w", err)
	}
	var model *idm.EnvelopeModel
	if i, ok := inv.(*Invocation); ok {
		model = i.Model()
	} else {
		var aud *did.DID
		if inv.Audience() != nil {
			a := inv.Audience().DID()
			aud = &a
		}
		var args idm.ArgsModel
		if inv.Arguments() != nil {
			if a, ok := inv.Arguments().(idm.CBORMarshalable); !ok {
				return nil, fmt.Errorf("invocation args do not implement CBOR marshalable")
			} else {
				args = idm.ArgsModel{Value: a}
			}
		}
		var meta *idm.MetaModel
		if inv.Metadata() != nil {
			if a, ok := inv.Metadata().(idm.CBORMarshalable); !ok {
				return nil, fmt.Errorf("invocation args do not implement CBOR marshalable")
			} else {
				meta = &idm.MetaModel{Value: a}
			}
		}

		model = &idm.EnvelopeModel{
			Signature: inv.Signature().Bytes(),
			SigPayload: idm.SigPayloadModel{
				Header: sigHeader,
				TokenPayload1_0_0_rc1: &idm.TokenPayloadModel1_0_0_rc1{
					Iss:   inv.Issuer().DID(),
					Sub:   inv.Subject().DID(),
					Aud:   aud,
					Cmd:   inv.Command(),
					Args:  args,
					Prf:   inv.Proofs(),
					Meta:  meta,
					Nonce: inv.Nonce(),
					Exp:   inv.Expiration(),
					Iat:   inv.IssuedAt(),
					Cause: inv.Cause(),
				},
			},
		}
	}

	buf := bytes.NewBuffer([]byte{})
	err = model.MarshalCBOR(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(data []byte) (*Invocation, error) {
	model := idm.EnvelopeModel{}
	err := model.UnmarshalCBOR(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("unmarshaling CBOR: %w", err)
	}
	header, err := varsig.Decode(model.SigPayload.Header)
	if err != nil {
		return nil, fmt.Errorf("decoding varsig header: %w", err)
	}
	sig := signature.NewSignature(header, model.Signature)
	return &Invocation{sig: sig, model: &model}, nil
}

func Invoke(iss ucan.Signer) (*Invocation, error) {
	if iss.SignatureCode() != ed25519.Code {
		return nil, fmt.Errorf("unknown signature code: %d", iss.SignatureCode())
	}
	h, err := varsig.Encode(common.Ed25519DagCbor)
	if err != nil {
		return nil, fmt.Errorf("encoding varsig header: %w", err)
	}

	tokenPayload := &idm.TokenPayloadModel1_0_0_rc1{}

	sigPayload := idm.SigPayloadModel{
		Header:                h,
		TokenPayload1_0_0_rc1: tokenPayload,
	}

	buf := bytes.NewBuffer([]byte{})
	err = tokenPayload.MarshalCBOR(buf)
	if err != nil {
		return nil, fmt.Errorf("marshaling token payload: %w", err)
	}

	sigBytes := iss.Sign(buf.Bytes())
	sig := signature.NewSignature(common.Ed25519DagCbor, sigBytes)

	model := idm.EnvelopeModel{
		Signature:  sigBytes,
		SigPayload: sigPayload,
	}

	return &Invocation{sig: sig, model: &model}, nil
}
