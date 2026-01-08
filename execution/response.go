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
	signer   ucan.Signer
	task     cid.Cid
	receipt  ucan.Receipt
	metadata ucan.Container
}

type ResponseOption func(r *ExecResponse) error

func WithSigner(signer ucan.Signer) ResponseOption {
	return func(resp *ExecResponse) error {
		resp.signer = signer
		return nil
	}
}

func WithReceipt(receipt ucan.Receipt) ResponseOption {
	return func(resp *ExecResponse) error {
		resp.SetReceipt(receipt)
		return nil
	}
}

// WithSuccess issues and sets a receipt for a successful execution of a task.
func WithSuccess(o ipld.Any) ResponseOption {
	return func(resp *ExecResponse) error {
		return resp.SetSuccess(o)
	}
}

// WithFailure issues and sets a receipt for a failed execution of a task.
func WithFailure(x error) ResponseOption {
	return func(resp *ExecResponse) error {
		return resp.SetFailure(x)
	}
}

func WithMetadata(m ucan.Container) ResponseOption {
	return func(resp *ExecResponse) error {
		resp.SetMetadata(m)
		return nil
	}
}

// NewResponse creates a new response object, representing the result of
// executing a task.
func NewResponse(task cid.Cid, options ...ResponseOption) (*ExecResponse, error) {
	response := ExecResponse{task: task}
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

func (r *ExecResponse) Receipt() ucan.Receipt {
	return r.receipt
}

func (r *ExecResponse) SetFailure(x error) error {
	if r.signer == nil {
		return fmt.Errorf("cannot issue receipt: missing signer")
	}
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
	receipt, err := receipt.Issue(r.signer, r.task, out)
	if err != nil {
		return err
	}
	r.receipt = receipt
	return nil
}

func (r *ExecResponse) SetMetadata(meta ucan.Container) error {
	r.metadata = meta
	return nil
}

func (r *ExecResponse) SetReceipt(receipt ucan.Receipt) error {
	if receipt.Ran() != r.task {
		return fmt.Errorf("cannot set receipt: task mismatch (expected %s, got %s)", r.task, receipt.Ran())
	}
	r.receipt = receipt
	return nil
}

func (r *ExecResponse) SetSigner(signer ucan.Signer) error {
	r.signer = signer
	return nil
}

func (r *ExecResponse) SetSuccess(o ipld.Any) error {
	if r.signer == nil {
		return fmt.Errorf("cannot issue receipt: missing signer")
	}
	out := result.OK[ipld.Any, ipld.Any](o)
	receipt, err := receipt.Issue(r.signer, r.task, out)
	if err != nil {
		return err
	}
	r.receipt = receipt
	return nil
}

var _ Response = (*ExecResponse)(nil)
