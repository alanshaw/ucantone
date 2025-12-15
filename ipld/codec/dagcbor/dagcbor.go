package dagcbor

import (
	cbg "github.com/whyrusleeping/cbor-gen"
)

const (
	Code        = 0x71
	ContentType = "application/vnd.ipld.dag-cbor"
)

type Marshaler = cbg.CBORMarshaler
type Unmarshaler = cbg.CBORUnmarshaler

type Marshalable interface {
	Marshaler
	Unmarshaler
}
