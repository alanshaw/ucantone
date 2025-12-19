package execution

import (
	"github.com/alanshaw/ucantone/errors"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
)

type ExecResponse struct {
	out      result.Result[ipld.Any, ipld.Any]
	metadata ucan.Container
}

type ResponseOption func(r *ExecResponse) error

func WithOutcome(out result.Result[ipld.Any, ipld.Any]) ResponseOption {
	return func(resp *ExecResponse) error {
		resp.out = out
		return nil
	}
}

func WithSuccess(o ipld.Any) ResponseOption {
	return func(resp *ExecResponse) error {
		resp.out = result.OK[ipld.Any, ipld.Any](o)
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
			name := "UnknownError"
			if nx, ok := x.(errors.Named); ok {
				name = nx.Name()
			}
			m["name"] = name
			m["message"] = x.Error()
		}
		resp.out = result.Error[ipld.Any, ipld.Any](ipld.Map(m))
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
	response := ExecResponse{out: result.OK[ipld.Any, ipld.Any](datamodel.Map{})}
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

func (r *ExecResponse) Out() result.Result[ipld.Any, ipld.Any] {
	return r.out
}
