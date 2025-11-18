package capability

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/invocation"
	verrs "github.com/alanshaw/ucantone/validator/errors"
	"github.com/ipfs/go-cid"
)

type Arguments interface {
	dagcbor.CBORMarshalable
}

type Task[A Arguments] struct {
	*invocation.Task
}

func NewTask[A Arguments](
	subject ucan.Subject,
	command ucan.Command,
	arguments ipld.Map,
	nonce ucan.Nonce,
) (*Task[A], error) {
	task, err := invocation.NewTask(subject, command, arguments, nonce)
	if err != nil {
		return nil, err
	}
	return &Task[A]{Task: task}, nil
}

// BindArguments binds the arguments to the arguments type for this task.
func (t *Task[A]) BindArguments() (A, error) {
	var args A
	// if args is a pointer type, then we need to create an instance of it because
	// rebind requires a non-nil pointer.
	typ := reflect.TypeOf(args)
	if typ.Kind() == reflect.Ptr {
		args = reflect.New(typ.Elem()).Interface().(A)
	}
	if err := datamodel.Rebind(datamodel.Map(t.Arguments()), args); err != nil {
		return args, verrs.NewMalformedArgumentsError(t.Command(), err)
	}
	return args, nil
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

func New[A Arguments](cmd ucan.Command, options ...Option) (*Capability[A], error) {
	cfg := capabilityConfig{pol: policy.Policy{}}
	for _, opt := range options {
		err := opt(&cfg)
		if err != nil {
			return nil, err
		}
	}
	cmd, err := command.Parse(string(cmd))
	if err != nil {
		return nil, fmt.Errorf("parsing command: %w", err)
	}
	return &Capability[A]{cmd, cfg.pol}, nil
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
			return nil, verrs.NewUnavailableProofError(p, errors.New("missing from map"))
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
	return invocation.Invoke(issuer, subject, c.cmd, m, options...)
}
