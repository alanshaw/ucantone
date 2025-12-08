package execution

import (
	"context"

	"github.com/alanshaw/ucantone/ucan"
)

type Request struct {
	ctx        context.Context
	invocation ucan.Invocation
	metadata   ucan.Container
}

func NewRequest(ctx context.Context, inv ucan.Invocation, meta ucan.Container) *Request {
	return &Request{
		ctx:        ctx,
		invocation: inv,
		metadata:   meta,
	}
}

func (r *Request) Context() context.Context {
	return r.ctx
}

func (r *Request) Invocation() ucan.Invocation {
	return r.invocation
}

func (r *Request) Metadata() ucan.Container {
	return r.metadata
}

func (r *Request) Task() ucan.Task {
	return r.invocation.Task()
}
