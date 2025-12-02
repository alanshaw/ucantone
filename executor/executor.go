package executor

import (
	"context"

	"github.com/alanshaw/ucantone/ipld"
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

// DispatchExecutor executes UCAN invocations by dispatching them to registered
// handlers.
type DispatchExecutor struct {
	handlers map[ucan.Command]handler
}

func New() *DispatchExecutor {
	return &DispatchExecutor{
		handlers: map[ucan.Command]handler{},
	}
}

func (d *DispatchExecutor) Handle(capability validator.Capability, fn HandlerFunc) {
	d.handlers[capability.Command()] = handler{Func: fn, Capability: capability}
}

func (d *DispatchExecutor) Execute(req ExecutionRequest, res ExecutionResponse) error {
	handler, ok := d.handlers[req.Task().Command()]
	if !ok {
		// TODO: transform into unknown command error
		panic("handler not found")
	}

	_, err := validator.Access(
		req.Context(),
		d.authority,
		handler.Capability,
		req.Invocation(),
		// TODO - options
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
