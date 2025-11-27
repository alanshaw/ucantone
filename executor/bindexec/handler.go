package bindexec

import (
	"context"

	"github.com/alanshaw/ucantone/executor"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ucan"
)

type Arguments interface {
	dagcbor.CBORUnmarshaler
}

type Success interface {
	dagcbor.CBORMarshaler
}

type Input[A Arguments] struct {
	executor.Input
	task *Task[A]
}

func (r *Input[A]) Task() *Task[A] {
	return r.task
}

type Output[O Success] struct {
	executor.Output
}

// SetResult sets the result of the task.
func (r *Output[O]) SetResult(ok O, err error) {
	r.Output.SetResult(ok, err)
}

// SetMetadata allows additional delegations, invocations and/or receipts to
// be sent in the response.
func (r *Output[O]) SetMetadata(metadata ucan.Container) {
	r.Output.SetMetadata(metadata)
}

type HandlerFunc[A Arguments, O Success] = func(ctx context.Context, request Input[A], response Output[O]) error

// NewHandler creates a new [executor.HandlerFunc] from the provided typed
// handler.
func NewHandler[A Arguments, O Success](handler HandlerFunc[A, O]) executor.HandlerFunc {
	return func(ctx context.Context, req executor.Input, res executor.Output) error {
		inv := req.Invocation()
		task, err := NewTask[A](inv.Subject(), inv.Command(), inv.Arguments(), inv.Nonce())
		if err != nil {
			// TODO: transform into malformed arguments error
			return err
		}
		ereq := Input[A]{Input: req, task: task}
		eres := Output[O]{Output: res}
		return handler(ctx, ereq, eres)
	}
}
