package execution

import (
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
)

type Response struct {
	result   result.Result[ipld.Any, ipld.Any]
	metadata ucan.Container
}

func NewResponse() *Response {
	return &Response{result: result.OK[ipld.Any, ipld.Any](datamodel.Map{})}
}

func (r *Response) Metadata() ucan.Container {
	return r.metadata
}

func (r *Response) Result() result.Result[ipld.Any, ipld.Any] {
	return r.result
}

func (r *Response) SetMetadata(meta ucan.Container) error {
	r.metadata = meta
	return nil
}

func (r *Response) SetResult(o ipld.Any, x error) error {
	if x != nil {
		m := datamodel.Map{}
		if cmx, ok := x.(dagcbor.CBORMarshaler); ok {
			err := datamodel.Rebind(cmx, &m)
			if err != nil {
				return err
			}
		} else {
			m["name"] = "UnknownError"
			m["message"] = x.Error()
		}
		r.result = result.Error[ipld.Any, ipld.Any](x)
	} else {
		r.result = result.OK[ipld.Any, ipld.Any](o)
	}
	return nil
}
