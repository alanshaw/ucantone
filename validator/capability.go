package validator

import (
	"errors"
	"fmt"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/ipfs/go-cid"
)

type Arguments interface {
	dagcbor.CBORMarshalable
}

type Task[A Arguments] struct {
	*invocation.Task
	args A
}

func NewTask[A Arguments](
	subject ucan.Subject,
	command ucan.Command,
	arguments ipld.Map[string, ipld.Any],
	nonce ucan.Nonce,
) (*Task[A], error) {
	cmargs, ok := arguments.(dagcbor.CBORMarshaler)
	if !ok {
		return nil, fmt.Errorf("arguments are not CBOR marshaler")
	}
	var args A
	if err := datamodel.Rebind(cmargs, args); err != nil {
		return nil, NewMalformedArgumentsError(command, err)
	}
	task, err := invocation.NewTask(subject, command, arguments, nonce)
	if err != nil {
		return nil, err
	}
	return &Task[A]{
		Task: task,
		args: args,
	}, nil
}

// BindArgs returns the arguments bound to the argument type var for this task.
func (t *Task[A]) BindArgs() A {
	return t.args
}

var _ ucan.Task = (*Task[Arguments])(nil)

type Match[A Arguments] struct {
	Invocation ucan.Invocation
	Value      *Task[A]
	Proofs     map[cid.Cid]ucan.Delegation
}

type Capability[A Arguments] struct {
	cmd ucan.Command
	pol ucan.Policy
}

func NewCapability[A Arguments](cmd ucan.Command, pol ucan.Policy) *Capability[A] {
	return &Capability[A]{cmd, pol}
}

// Match an invocation against the capability, resulting in a match, which is
// the task from the invocation, verified to be matching with delegation
// policies.
func (c *Capability[A]) Match(inv ucan.Invocation, proofs map[cid.Cid]ucan.Delegation) (*Match[A], error) {
	ok, err := policy.Match(c.pol, inv.Arguments())
	if !ok {
		return nil, err
	}

	for _, p := range inv.Proofs() {
		prf, ok := proofs[p]
		if !ok {
			return nil, NewUnavailableProofError(p, errors.New("missing from map"))
		}
		ok, err = policy.Match(prf.Policy(), inv.Arguments())
		if !ok {
			return nil, err
		}
	}

	task, err := NewTask[A](inv.Subject(), inv.Command(), inv.Arguments(), inv.Nonce())
	if err != nil {
		return nil, err
	}

	return &Match[A]{Invocation: inv, Value: task, Proofs: proofs}, nil
}

func (c *Capability[A]) Command() ucan.Command {
	return c.cmd
}

func (c *Capability[A]) Policy() ucan.Policy {
	return c.pol
}

func (c *Capability[A]) Delegate(issuer ucan.Signer, audience ucan.Principal, options ...delegation.Option) (*delegation.Delegation, error) {
	return delegation.Delegate(issuer, audience, c.cmd, options...)
}

func (c *Capability[A]) Invoke(issuer ucan.Signer, subject ucan.Subject, arguments A, options ...invocation.Option) (*invocation.Invocation, error) {
	var m datamodel.Map
	err := datamodel.Rebind(arguments, &m)
	if err != nil {
		return nil, err
	}
	return invocation.Invoke(issuer, subject, c.cmd, &m)
}
