package dagcbor

import (
	cbg "github.com/whyrusleeping/cbor-gen"
)

const Code = 0x71

type CBORMarshaler = cbg.CBORMarshaler
type CBORUnmarshaler = cbg.CBORUnmarshaler
