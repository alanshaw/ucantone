package ipld

import (
	"github.com/ipfs/go-cid"
)

// Any is an alias for any/interface{}, however it denotes that the type MUST
// be an IPLD type. i.e. one of:
//
//   - Null (nil)
//   - Boolean (bool)
//   - Integer (int64, int)
//   - String (string)
//   - Bytes ([]byte)
//   - List (slice)
//   - Map ([Map])
//   - Link ([cid.Cid])
type Any = any

type Map = map[string]Any

// Block is content addressed and encoded IPLD data.
type Block interface {
	Link() cid.Cid
	Bytes() []byte
}
