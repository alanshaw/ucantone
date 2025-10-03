package ipld

import (
	"iter"
)

type Map[K, V any] interface {
	// Keys gets the list of keys in this map.
	Keys() iter.Seq[K]
	// Get a value for the given key. It returns false if the key does not exist
	// in the map.
	Get(k K) (V, bool)
	// Set a value for the given key.
	Set(k K, v V)
}
