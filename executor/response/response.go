package response

import (
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ucan"
)

type Response struct {
	ok  ipld.Any
	err error
	ct  ucan.Container
}

func New() *Response {
	return &Response{}
}

func (r *Response) SetContainer(ct ucan.Container) {
	r.ct = ct
}

func (r *Response) SetResult(ok ipld.Any, err error) {
	r.ok = ok
	r.err = err
}
