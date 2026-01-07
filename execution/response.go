package execution

import (
	"fmt"

	"github.com/alanshaw/ucantone/errors"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/receipt"
	"github.com/ipfs/go-cid"
)

type ExecResponse struct {
	receipt  ucan.Receipt
	metadata ucan.Container
}

type ResponseOption func(r *ExecResponse) error

func WithReceipt(receipt ucan.Receipt) ResponseOption {
	return func(resp *ExecResponse) error {
		resp.receipt = receipt
		return nil
	}
}

// WithSuccess issues and sets a receipt for a successful execution of a task.
func WithSuccess(signer ucan.Signer, task cid.Cid, o ipld.Any) ResponseOption {
	return func(resp *ExecResponse) error {
		out := result.OK[ipld.Any, ipld.Any](o)
		receipt, err := receipt.Issue(signer, task, out)
		if err != nil {
			return err
		}
		resp.receipt = receipt
		return nil
	}
}

// WithFailure issues and sets a receipt for a failed execution of a task.
func WithFailure(signer ucan.Signer, task cid.Cid, x error) ResponseOption {
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
		out := result.Error[ipld.Any, ipld.Any](ipld.Map(m))
		receipt, err := receipt.Issue(signer, task, out)
		if err != nil {
			return err
		}
		resp.receipt = receipt
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
	response := ExecResponse{}
	for _, opt := range options {
		err := opt(&response)
		if err != nil {
			return nil, err
		}
	}
	if response.receipt == nil {
		return nil, fmt.Errorf("missing response receipt") // developer error
	}
	return &response, nil
}

func (r *ExecResponse) Metadata() ucan.Container {
	return r.metadata
}

func (r *ExecResponse) Receipt() ucan.Receipt {
	return r.receipt
}
