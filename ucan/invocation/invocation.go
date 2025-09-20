package invocation

import (
	"bytes"

	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
)

type Invocation struct {
	model *idm.InvocationModel
}

func (inv Invocation) Model() *idm.InvocationModel {
	return inv.model
}

func Encode(inv Invocation) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := inv.Model().MarshalCBOR(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(data []byte) (Invocation, error) {
	model := idm.InvocationModel{}
	err := model.UnmarshalCBOR(bytes.NewReader(data))
	if err != nil {
		return Invocation{}, err
	}
	return Invocation{model: &model}, nil
}
