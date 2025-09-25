package ipld

import "iter"

type Map[K, V any] interface {
	Keys() iter.Seq[K]
	Value(k K) V
}
