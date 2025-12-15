package execution

import (
	"context"

	"github.com/alanshaw/ucantone/ucan"
)

type ExecRequest struct {
	ctx        context.Context
	invocation ucan.Invocation
	metadata   ucan.Container
}

func NewRequest(ctx context.Context, inv ucan.Invocation, meta ucan.Container) *ExecRequest {
	return &ExecRequest{
		ctx:        ctx,
		invocation: inv,
		metadata:   meta,
	}
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
