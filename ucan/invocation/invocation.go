package invocation

import (
	"bytes"
	"fmt"

	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/common"
)

type Invocation struct {
	model *idm.EnvelopeModel
}

func (inv *Invocation) Model() *idm.EnvelopeModel {
	return inv.model
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

func Invoke() (*Invocation, error) {
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
