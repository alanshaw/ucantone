package dispatcher

import (
	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/validator"
)

type handler struct {
	Func       execution.HandlerFunc
	Capability validator.Capability
}

// Dispatcher executes UCAN invocations by dispatching them to registered
// handlers.
type Dispatcher struct {
	authority      ucan.Verifier
	handlers       map[ucan.Command]handler
	validationOpts []validator.Option
}

// New creates an invocation executor that executes UCAN invocations by
// dispatching them to registered handlers.
//
// The authority is the identity of the local authority, used to verify
// signatures of delegations signed by it.
func New(authority ucan.Verifier, options ...Option) *Dispatcher {
	cfg := execConfig{}
	for _, opt := range options {
		opt(&cfg)
	}
	return &Dispatcher{
		authority:      authority,
		handlers:       map[ucan.Command]handler{},
		validationOpts: cfg.validationOpts,
	}
}

func (d *Dispatcher) Handle(capability validator.Capability, fn execution.HandlerFunc) {
	d.handlers[capability.Command()] = handler{Func: fn, Capability: capability}
}

func (d *Dispatcher) Execute(req execution.Request) (execution.Response, error) {
	aud := req.Invocation().Audience()
	if aud.DID() != d.authority.DID() {
		return execution.NewResponse(
			execution.WithFailure(execution.NewInvalidAudienceError(d.authority, aud)),
		)
	}

	cmd := req.Invocation().Command()
	handler, ok := d.handlers[cmd]
	if !ok {
		return execution.NewResponse(execution.WithFailure(NewHandlerNotFoundError(cmd)))
	}

	_, err := validator.Access(
		req.Context(),
		d.authority,
		handler.Capability,
		req.Invocation(),
		d.validationOpts...,
	)
	if err != nil {
		return execution.NewResponse(execution.WithFailure(err))
	}

	res, err := handler.Func(req)
	if err != nil {
		return execution.NewResponse(
			execution.WithFailure(execution.NewHandlerExecutionError(cmd, err)),
		)
	}
	return res, nil
}
