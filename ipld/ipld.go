package ipld

import (
	"iter"
)

// Any is an alias for any/interface{}, however it denotes that the type must
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

type Map[K, V Any] interface {
	// Keys gets the list of keys in this map.
	Keys() iter.Seq[K]
	// Get a value for the given key. It returns false if the key does not exist
	// in the map.
	Get(k K) (V, bool)
	// Set a value for the given key.
	Set(k K, v V)
}
