package dagcbor

import cbg "github.com/whyrusleeping/cbor-gen"

// CBORMarshalable is an interface describing a type that allows both
// marshalling to CBOR as well as unmarshalling from CBOR.
type CBORMarshalable interface {
	cbg.CBORMarshaler
	cbg.CBORUnmarshaler
}

type CBORMarshaler = cbg.CBORMarshaler
type CBORUnmarshaler = cbg.CBORUnmarshaler
