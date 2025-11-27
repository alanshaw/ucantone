package request

import (
	"github.com/alanshaw/ucantone/ucan"
)

type Request struct {
	ct  ucan.Container
	inv ucan.Invocation
}

func New(container ucan.Container, invocation ucan.Invocation) *Request {
	return &Request{
		ct:  container,
		inv: invocation,
	}
}

func (r *Request) Container() ucan.Container {
	return r.ct
}

func (r *Request) Invocation() ucan.Invocation {
	return r.inv
}

func (r *Request) Task() ucan.Task {
	return r.inv.Task()
}
