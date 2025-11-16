package ipld

import (
	"iter"

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

// Map is an IPLD map that supports any of the IPLD types for values. Keys MUST
// be strings.
type Map interface {
	// All is an iterator over all the key/value pairs in this map.
	All() iter.Seq2[string, Any]
	// Get a value for the given key. It returns false if the key does not exist
	// in the map.
	Get(k string) (Any, bool)
	// Keys is an iterator over the keys in this map.
	Keys() iter.Seq[string]
	// Values is an iterator over the values in this map.
	Values() iter.Seq[Any]
}

// MutableMap is a [Map] that supports changes.
type MutableMap interface {
	// Set a value for the given key.
	Set(k string, v Any)
}

// Block is content addressed and encoded IPLD data.
type Block interface {
	Link() cid.Cid
	Bytes() []byte
}
