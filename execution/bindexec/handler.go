package bindexec

import (
	"context"

	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/ipfs/go-cid"
)

type Arguments interface {
	dagcbor.Unmarshaler
}

type Success interface {
	dagcbor.Marshaler
}

type requestConfig struct {
	invocations []ucan.Invocation
	delegations []ucan.Delegation
	receipts    []ucan.Receipt
}

type RequestOption = func(cfg *requestConfig)

// WithProofs adds delegations to the execution request. They should be linked
// from the invocation to be executed.
func WithProofs(delegations ...ucan.Delegation) RequestOption {
	return func(cfg *requestConfig) {
		cfg.delegations = append(cfg.delegations, delegations...)
	}
}

// WithDelegations adds delegations to the execution request.
func WithDelegations(delegations ...ucan.Delegation) RequestOption {
	return func(cfg *requestConfig) {
		cfg.delegations = append(cfg.delegations, delegations...)
	}
}

// WithReceipts adds receipts to the execution request.
func WithReceipts(receipts ...ucan.Receipt) RequestOption {
	return func(cfg *requestConfig) {
		cfg.receipts = append(cfg.receipts, receipts...)
	}
}

// WithInvocations adds additional invocations to the execution request.
func WithInvocations(invocations ...ucan.Invocation) RequestOption {
	return func(cfg *requestConfig) {
		cfg.invocations = append(cfg.invocations, invocations...)
	}
}

type Request[A Arguments] struct {
	execution.Request
	task *Task[A]
}

func NewRequest[A Arguments](ctx context.Context, inv ucan.Invocation, options ...RequestOption) (*Request[A], error) {
	cfg := requestConfig{}
	for _, opt := range options {
		opt(&cfg)
	}
	task, err := NewTask[A](inv.Subject(), inv.Command(), inv.Arguments(), inv.Nonce())
	if err != nil {
		return nil, err
	}
	return &Request[A]{
		Request: execution.NewRequest(
			ctx,
			inv,
			execution.WithInvocations(cfg.invocations...),
			execution.WithDelegations(cfg.delegations...),
			execution.WithReceipts(cfg.receipts...),
		),
		task: task,
	}, nil
}

// Task returns an object containing just the fields that comprise the task
// for the invocation.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#task
func (r *Request[A]) Task() *Task[A] {
	return r.task
}

type ResponseOption[O Success] func(r *Response[O]) error

func WithReceipt[O Success](receipt ucan.Receipt) ResponseOption[O] {
	return func(resp *Response[O]) error {
		exr, err := execution.NewResponse(
			execution.WithReceipt(receipt),
			execution.WithMetadata(resp.Response.Metadata()),
		)
		if err != nil {
			return err
		}
		resp.Response = exr
		return nil
	}
}

// WithSuccess issues and sets a receipt for a successful execution of a task.
func WithSuccess[O Success](signer ucan.Signer, task cid.Cid, o O) ResponseOption[O] {
	return func(resp *Response[O]) error {
		m := datamodel.Map{}
		err := datamodel.Rebind(o, &m)
		if err != nil {
			return err
		}
		exr, err := execution.NewResponse(
			execution.WithSuccess(signer, task, m),
			execution.WithMetadata(resp.Metadata()),
		)
		if err != nil {
			return err
		}
		resp.Response = exr
		return nil
	}
}

// WithFailure issues and sets a receipt for a failed execution of a task.
func WithFailure[O Success](signer ucan.Signer, task cid.Cid, x error) ResponseOption[O] {
	return func(resp *Response[O]) error {
		exr, err := execution.NewResponse(
			execution.WithFailure(signer, task, x),
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
	return func(resp *Response[O]) error {
		exr, err := execution.NewResponse(
			execution.WithReceipt(resp.Receipt()),
			execution.WithMetadata(m),
		)
		if err != nil {
			return err
		}
		resp.Response = exr
		return nil
	}
}

type Response[O Success] struct {
	execution.Response
}

func NewResponse[O Success](options ...ResponseOption[O]) (*Response[O], error) {
	response := Response[O]{Response: &execution.ExecResponse{}}
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
			return nil, NewMalformedArgumentsError(err)
		}
		return handler(&Request[A]{Request: req, task: task})
	}
}
