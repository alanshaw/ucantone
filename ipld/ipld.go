package ipld

import (
	"iter"
)

type Map[K, V any] interface {
	// Keys gets the list of keys in this map.
	Keys() iter.Seq[K]
	// Value gets a value for the given key. It returns false if the key does not
	// exist in the map.
	Value(k K) (V, bool)
}
