package execution

import (
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
)

type ExecResponse struct {
	result   result.Result[ipld.Any, ipld.Any]
	metadata ucan.Container
}

type ResponseOption func(r *ExecResponse) error

func WithResult(r result.Result[ipld.Any, ipld.Any]) ResponseOption {
	return func(resp *ExecResponse) error {
		resp.result = r
		return nil
	}
}

func WithSuccess(o ipld.Any) ResponseOption {
	return func(resp *ExecResponse) error {
		resp.result = result.OK[ipld.Any, ipld.Any](o)
		return nil
	}
}

func WithFailure(x error) ResponseOption {
	return func(resp *ExecResponse) error {
		m := datamodel.Map{}
		if cmx, ok := x.(dagcbor.Marshaler); ok {
			err := datamodel.Rebind(cmx, &m)
			if err != nil {
				return err
			}
		} else {
			m["name"] = "UnknownError"
			m["message"] = x.Error()
		}
		resp.result = result.Error[ipld.Any, ipld.Any](x)
		return nil
	}
}

func WithMetadata(m ucan.Container) ResponseOption {
	return func(r *ExecResponse) error {
		r.metadata = m
		return nil
	}
}

func NewResponse(options ...ResponseOption) (*ExecResponse, error) {
	response := ExecResponse{result: result.OK[ipld.Any, ipld.Any](datamodel.Map{})}
	for _, opt := range options {
		err := opt(&response)
		if err != nil {
			return nil, err
		}
	}
	return &response, nil
}

func (r *ExecResponse) Metadata() ucan.Container {
	return r.metadata
}

func (r *ExecResponse) Result() result.Result[ipld.Any, ipld.Any] {
	return r.result
}
