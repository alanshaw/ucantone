package dagcbor

import (
	cbg "github.com/whyrusleeping/cbor-gen"
)

const (
	Code        = 0x71
	ContentType = "application/vnd.ipld.dag-cbor"
)

type CBORMarshaler = cbg.CBORMarshaler
type CBORUnmarshaler = cbg.CBORUnmarshaler

type CBORMarshalable interface {
	CBORMarshaler
	CBORUnmarshaler
}
