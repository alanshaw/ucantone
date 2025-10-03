package datamodel

import (
	"bytes"
	"errors"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
)

// Bind binds the passed IPLD map to the passed Go struct.
func Bind(m ipld.Map[string, any], ptr dagcbor.CBORUnmarshaler) error {
	cm, ok := m.(dagcbor.CBORMarshaler)
	if !ok {
		return errors.New("map is not a CBOR marshaler")
	}
	var buf bytes.Buffer
	if err := cm.MarshalCBOR(&buf); err != nil {
		return err
	}
	return ptr.UnmarshalCBOR(&buf)
}
