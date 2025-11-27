package executor

import (
	"context"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ucan"
)

type Input interface {
	// Task from the invocation that should be performed.
	Task() ucan.Task
	// Invocation that should be executed.
	Invocation() ucan.Invocation
	// Metadata provides additional information about the invocation.
	Metadata() ucan.Container
}

type Output interface {
	// SetResult sets the result of the task.
	SetResult(ipld.Any, error)
	// SetMetadata allows additional delegations, invocations and/or receipts to
	// be sent in the response.
	SetMetadata(ucan.Container)
}

type HandlerFunc = func(ctx context.Context, in Input, out Output) error

// Executor executes UCAN invocations by dispatching them to registered
// handlers. It also uses the validator to handle invocation proof chain
// validation and policy matching.
type Executor struct {
	handlers map[ucan.Command]HandlerFunc
}

func New() *Executor {
	return &Executor{
		handlers: map[ucan.Command]HandlerFunc{},
	}
}

func (e *Executor) Handle(cmd ucan.Command, handler HandlerFunc) {
	e.handlers[cmd] = handler
}

func (e *Executor) Execute(ctx context.Context, in Input, out Output) {
	handler, ok := e.handlers[in.Task().Command()]
	if !ok {
		// TODO: transform into unknown command error
		panic("handler not found")
	}

	err := handler(ctx, in, out)
	if err != nil {
		// TODO: transform into handler execution error
		out.SetResult(nil, err)
	}
}
