package execution

import (
	"context"

	"github.com/alanshaw/ucantone/ucan"
)

type RequestOption = func(r *ExecRequest)

func WithRequestMetadata(meta ucan.Container) RequestOption {
	return func(r *ExecRequest) {
		r.metadata = meta
	}
}

type ExecRequest struct {
	ctx        context.Context
	invocation ucan.Invocation
	metadata   ucan.Container
}

func NewRequest(ctx context.Context, inv ucan.Invocation, options ...RequestOption) *ExecRequest {
	req := &ExecRequest{
		ctx:        ctx,
		invocation: inv,
	}
	for _, opt := range options {
		opt(req)
	}
	return req
}

func (r *ExecRequest) Context() context.Context {
	return r.ctx
}

func (r *ExecRequest) Invocation() ucan.Invocation {
	return r.invocation
}

func (r *ExecRequest) Metadata() ucan.Container {
	return r.metadata
}
