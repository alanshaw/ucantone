package datamodel

import (
	"bytes"

	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
)

// Rebind binds the data from one CBOR marshaler type to another CBOR
// unmarshaler type.
func Rebind[A dagcbor.CBORMarshaler, B dagcbor.CBORUnmarshaler](from A, to B) error {
	var buf bytes.Buffer
	if err := from.MarshalCBOR(&buf); err != nil {
		return err
	}
	return to.UnmarshalCBOR(&buf)
}
