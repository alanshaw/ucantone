package invocation

import (
	"io"
	"iter"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
)

// NoArguments can be used to issue an invocation with no arguments.
type NoArguments struct{}

func (n NoArguments) All() iter.Seq2[string, ipld.Any] {
	return func(func(string, ipld.Any) bool) {}
}

func (n NoArguments) Get(k string) (ipld.Any, bool) {
	return nil, false
}

func (n NoArguments) Keys() iter.Seq[string] {
	return func(func(string) bool) {}
}

func (n NoArguments) MarshalCBOR(w io.Writer) error {
	return datamodel.NewMap().MarshalCBOR(w)
}

func (n NoArguments) UnmarshalCBOR(r io.Reader) error {
	return datamodel.NewMap().UnmarshalCBOR(r)
}

func (n NoArguments) Values() iter.Seq[ipld.Any] {
	return func(func(ipld.Any) bool) {}
}

var _ ipld.Map[string, ipld.Any] = (*NoArguments)(nil)
var _ dagcbor.CBORMarshalable = (*NoArguments)(nil)

// UnknownArguments can be used when the arguments for an invocation cannot be
// bound to a known type.
type UnknownArguments = *datamodel.Map
