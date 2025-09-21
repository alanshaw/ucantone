package invocation

import (
	"bytes"
	"fmt"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan"
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/algoithm/ed25519"
	"github.com/alanshaw/ucantone/varsig/common"
)

type Invocation struct {
	model *idm.EnvelopeModel
}

func (inv *Invocation) Model() *idm.EnvelopeModel {
	return inv.model
}

func (inv *Invocation) Issuer() did.DID {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Iss
}

func (inv *Invocation) Subject() did.DID {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Sub
}

func (inv *Invocation) Audience() *did.DID {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Aud
}

func (inv *Invocation) Command() string {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Cmd
}

func Encode(inv *Invocation) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := inv.Model().MarshalCBOR(buf)
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
	_, err = varsig.Decode(model.SigPayload.Header)
	if err != nil {
		return nil, fmt.Errorf("decoding varsig header: %w", err)
	}
	return &Invocation{model: &model}, nil
}

func Invoke(iss ucan.Signer) (*Invocation, error) {
	if iss.SignatureCode() != ed25519.Code {
		return nil, fmt.Errorf("unknown signature code: %d", iss.SignatureCode())
	}
	h, err := varsig.Encode(common.Ed25519DagCbor)
	if err != nil {
		return nil, fmt.Errorf("encoding varsig header: %w", err)
	}

	sigPayload := idm.SigPayloadModel{
		Header:                h,
		TokenPayload1_0_0_rc1: &idm.TokenPayloadModel1_0_0_rc1{},
	}

	sig := []byte{} // TODO

	model := idm.EnvelopeModel{
		Signature:  sig,
		SigPayload: sigPayload,
	}

	return &Invocation{&model}, nil
}
