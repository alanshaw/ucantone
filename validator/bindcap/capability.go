package bindcap

import (
	"reflect"

	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/alanshaw/ucantone/validator/capability"
	verrs "github.com/alanshaw/ucantone/validator/errors"
	"github.com/ipfs/go-cid"
)

type Arguments interface {
	dagcbor.Marshalable
}

// Capability that can be used to validate an invocation against proof policies.
type Capability[A Arguments] struct {
	cap *capability.Capability
}

// New creates a new capability definition that can be used to validate an
// invocation against proof policies.
func New[A Arguments](cmd ucan.Command, options ...capability.Option) (*Capability[A], error) {
	cap, err := capability.New(cmd, options...)
	if err != nil {
		return nil, err
	}
	return &Capability[A]{cap}, nil
}

// Match an invocation against the capability, resulting in a match, which is
// the task from the invocation, verified to be matching with delegation
// policies.
//
// Match also ensures the invocation arguments can be bound to the
// specified Arguments type A.
func (c *Capability[A]) Match(inv ucan.Invocation, proofs map[cid.Cid]ucan.Delegation) (*capability.Match, error) {
	var args A
	// if args is a pointer type, then we need to create an instance of it because
	// rebind requires a non-nil pointer.
	typ := reflect.TypeOf(args)
	if typ.Kind() == reflect.Ptr {
		args = reflect.New(typ.Elem()).Interface().(A)
	}
	if err := datamodel.Rebind(datamodel.Map(inv.Arguments()), args); err != nil {
		return nil, verrs.NewMalformedArgumentsError(inv.Command(), err)
	}
	return c.cap.Match(inv, proofs)
}

func (c *Capability[A]) Command() ucan.Command {
	return c.cap.Command()
}

func (c *Capability[A]) Policy() ucan.Policy {
	return c.cap.Policy()
}

func (c *Capability[A]) Delegate(issuer ucan.Signer, audience ucan.Principal, subject ucan.Subject, options ...delegation.Option) (*delegation.Delegation, error) {
	return delegation.Delegate(issuer, audience, subject, c.cap.Command(), options...)
}

func (c *Capability[A]) Invoke(issuer ucan.Signer, subject ucan.Subject, arguments A, options ...invocation.Option) (*invocation.Invocation, error) {
	var args datamodel.Map
	err := datamodel.Rebind(arguments, &args)
	if err != nil {
		return nil, err
	}
	return invocation.Invoke(issuer, subject, c.cap.Command(), args, options...)
}
