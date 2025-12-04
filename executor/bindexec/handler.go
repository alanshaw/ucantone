package bindexec

import (
	"github.com/alanshaw/ucantone/executor"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/ucan"
)

type Arguments interface {
	dagcbor.CBORUnmarshaler
}

type Success interface {
	dagcbor.CBORMarshaler
}

type ExecutionRequest[A Arguments] struct {
	executor.ExecutionRequest
	task *Task[A]
}

func (r *ExecutionRequest[A]) Task() *Task[A] {
	return r.task
}

type ExecutionResponse[O Success] struct {
	executor.ExecutionResponse
}

// SetResult sets the result of the task.
func (r *ExecutionResponse[O]) SetResult(o O, x error) error {
	if x == nil {
		m := datamodel.Map{}
		err := datamodel.Rebind(o, &m)
		if err != nil {
			return err
		}
		return r.ExecutionResponse.SetResult(m, nil)
	}
	return r.ExecutionResponse.SetResult(nil, x)
}

// SetMetadata allows additional delegations, invocations and/or receipts to
// be sent in the response.
func (r *ExecutionResponse[O]) SetMetadata(metadata ucan.Container) error {
	return r.ExecutionResponse.SetMetadata(metadata)
}

type HandlerFunc[A Arguments, O Success] = func(req ExecutionRequest[A], res ExecutionResponse[O]) error

// NewHandler creates a new [executor.HandlerFunc] from the provided typed
// handler.
func NewHandler[A Arguments, O Success](handler HandlerFunc[A, O]) executor.HandlerFunc {
	return func(req executor.ExecutionRequest, res executor.ExecutionResponse) error {
		inv := req.Invocation()
		task, err := NewTask[A](inv.Subject(), inv.Command(), inv.Arguments(), inv.Nonce())
		if err != nil {
			// TODO: transform into malformed arguments error
			return err
		}
		return handler(
			ExecutionRequest[A]{ExecutionRequest: req, task: task},
			ExecutionResponse[O]{ExecutionResponse: res},
		)
	}
}
