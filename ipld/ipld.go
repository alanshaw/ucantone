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

// Map is an IPLD map that supports any of the IPLD types for keys/values.
type Map[K, V Any] interface {
	// Entries is an iterator over the key/value pairs in this map.
	Entries() iter.Seq2[K, V]
	// Get a value for the given key. It returns false if the key does not exist
	// in the map.
	Get(k K) (V, bool)
	// Keys is an iterator over the keys in this map.
	Keys() iter.Seq[K]
	// Values is an iterator over the values in this map.
	Values() iter.Seq[V]
}

// MutableMap is a [Map] that supports changes.
type MutableMap[K, V Any] interface {
	// Set a value for the given key.
	Set(k K, v V)
}

// Block is content addressed and encoded IPLD data.
type Block interface {
	Link() cid.Cid
	Bytes() []byte
}
