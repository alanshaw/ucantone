package invocation

import (
	"bytes"
	"fmt"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/crypto/signature"
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/algoithm/ed25519"
	"github.com/alanshaw/ucantone/varsig/common"
	cid "github.com/ipfs/go-cid"
)

type Invocation struct {
	sig   *signature.Signature
	model *idm.EnvelopeModel
}

func (inv *Invocation) Args() any {
	panic("unimplemented")
}

func (inv *Invocation) Audience() ucan.Principal {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Aud
}

func (inv *Invocation) Cause() *cid.Cid {
	panic("unimplemented")
}

func (inv *Invocation) Command() ucan.Command {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Cmd
}

func (inv *Invocation) Expiration() *ucan.UTCUnixTimestamp {
	panic("unimplemented")
}

func (inv *Invocation) IssuedAt() *ucan.UTCUnixTimestamp {
	panic("unimplemented")
}

func (inv *Invocation) Issuer() ucan.Principal {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Iss
}

func (inv *Invocation) Metadata() any {
	panic("unimplemented")
}

func (inv *Invocation) Model() *idm.EnvelopeModel {
	return inv.model
}

func (inv *Invocation) Nonce() ucan.Nonce {
	panic("unimplemented")
}

func (inv *Invocation) Proofs() []cid.Cid {
	panic("unimplemented")
}

func (inv *Invocation) Signature() ucan.Signature {
	return inv.sig
}

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
		var args *idm.ArgsModel
		if inv.Args() != nil {
			if a, ok := inv.Args().(idm.CBORMarshalable); !ok {
				return nil, fmt.Errorf("invocation args do not implement CBOR marshalable")
			} else {
				args = &idm.ArgsModel{Value: a}
			}
		}
		var meta *idm.MetaModel
		if inv.Metadata() != nil {
			if a, ok := inv.Metadata().(idm.CBORMarshalable); !ok {
				return nil, fmt.Errorf("invocation args do not implement CBOR marshalable")
			} else {
				args = &idm.MetaModel{Value: a}
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
