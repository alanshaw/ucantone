package executor

import (
	"context"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/principal"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/validator"
)

type ExecutionRequest interface {
	Context() context.Context
	// Task from the invocation that should be performed.
	Task() ucan.Task
	// Invocation that should be executed.
	Invocation() ucan.Invocation
	// Metadata provides additional information about the invocation.
	Metadata() ucan.Container
}

type ExecutionResponse interface {
	// SetResult sets the result of the task.
	SetResult(ipld.Any, error) error
	// SetMetadata allows additional delegations, invocations and/or receipts to
	// be sent in the response.
	SetMetadata(ucan.Container) error
}

// Executor executes UCAN invocations. It also validates proof chains and
// matches policies.
type Executor interface {
	Execute(req ExecutionRequest, res ExecutionResponse) error
}

type HandlerFunc = func(req ExecutionRequest, res ExecutionResponse) error

type handler struct {
	Func       HandlerFunc
	Capability validator.Capability
}

// DispatchingExecutor executes UCAN invocations by dispatching them to registered
// handlers.
type DispatchingExecutor struct {
	id             principal.Signer
	handlers       map[ucan.Command]handler
	validationOpts []validator.Option
}

func New(id principal.Signer, options ...Option) *DispatchingExecutor {
	cfg := execConfig{}
	for _, opt := range options {
		opt(&cfg)
	}
	return &DispatchingExecutor{
		id:             id,
		handlers:       map[ucan.Command]handler{},
		validationOpts: cfg.validationOpts,
	}
}

func (d *DispatchingExecutor) Handle(capability validator.Capability, fn HandlerFunc) {
	d.handlers[capability.Command()] = handler{Func: fn, Capability: capability}
}

func (d *DispatchingExecutor) Execute(req ExecutionRequest, res ExecutionResponse) error {
	handler, ok := d.handlers[req.Task().Command()]
	if !ok {
		// TODO: transform into unknown command error
		panic("handler not found")
	}

	_, err := validator.Access(
		req.Context(),
		d.id.Verifier(),
		handler.Capability,
		req.Invocation(),
		d.validationOpts...,
	)
	if err != nil {
		return res.SetResult(nil, err)
	}

	err = handler.Func(req, res)
	if err != nil {
		// TODO: transform into handler execution error
		return res.SetResult(nil, err)
	}

	return nil
}
