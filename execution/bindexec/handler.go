package bindexec

import (
	"context"

	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
)

type Arguments interface {
	dagcbor.Unmarshaler
}

type Success interface {
	dagcbor.Marshaler
}

type requestConfig struct {
	metadata ucan.Container
}

type RequestOption = func(cfg *requestConfig)

func WithRequestMetadata(meta ucan.Container) RequestOption {
	return func(cfg *requestConfig) {
		cfg.metadata = meta
	}
}

type Request[A Arguments] struct {
	execution.Request
	task *Task[A]
}

func NewRequest[A Arguments](ctx context.Context, inv ucan.Invocation, options ...RequestOption) Request[A] {
	cfg := requestConfig{}
	for _, opt := range options {
		opt(&cfg)
	}
	return Request[A]{Request: execution.NewRequest(ctx, inv, execution.WithRequestMetadata(cfg.metadata))}
}

func (r *Request[A]) Task() *Task[A] {
	return r.task
}

type ResponseOption[O Success] func(r *Response[O]) error

func WithResult[O Success](r result.Result[O, error]) ResponseOption[O] {
	return func(resp *Response[O]) error {
		o, x := result.Unwrap(r)
		if x == nil {
			return WithSuccess(o)(resp)
		}
		return WithFailure[O](x)(resp)
	}
}

func WithSuccess[O Success](o O) ResponseOption[O] {
	return func(r *Response[O]) error {
		m := datamodel.Map{}
		err := datamodel.Rebind(o, &m)
		if err != nil {
			return err
		}
		exr, err := execution.NewResponse(
			execution.WithSuccess(m),
			execution.WithMetadata(r.Response.Metadata()),
		)
		if err != nil {
			return err
		}
		r.Response = exr
		return nil
	}
}

func WithFailure[O Success](x error) ResponseOption[O] {
	return func(resp *Response[O]) error {
		exr, err := execution.NewResponse(
			execution.WithFailure(x),
			execution.WithMetadata(resp.Response.Metadata()),
		)
		if err != nil {
			return err
		}
		resp.Response = exr
		return nil
	}
}

func WithMetadata[O Success](m ucan.Container) ResponseOption[O] {
	return func(r *Response[O]) error {
		exr, err := execution.NewResponse(
			execution.WithResult(r.Response.Result()),
			execution.WithMetadata(m),
		)
		if err != nil {
			return err
		}
		r.Response = exr
		return nil
	}
}

type Response[O Success] struct {
	execution.Response
}

func NewResponse[O Success](options ...ResponseOption[O]) (*Response[O], error) {
	exr, err := execution.NewResponse()
	if err != nil {
		return nil, err
	}
	response := Response[O]{Response: exr}
	for _, opt := range options {
		err := opt(&response)
		if err != nil {
			return nil, err
		}
	}
	return &response, nil
}

type HandlerFunc[A Arguments, O Success] = func(req *Request[A]) (*Response[O], error)

// NewHandler creates a new [execution.HandlerFunc] from the provided typed
// handler.
func NewHandler[A Arguments, O Success](handler HandlerFunc[A, O]) execution.HandlerFunc {
	return func(req execution.Request) (execution.Response, error) {
		inv := req.Invocation()
		task, err := NewTask[A](inv.Subject(), inv.Command(), inv.Arguments(), inv.Nonce())
		if err != nil {
			return execution.NewResponse(execution.WithFailure(NewMalformedArgumentsError(err)))
		}
		return handler(&Request[A]{Request: req, task: task})
	}
}
