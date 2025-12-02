package server

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/alanshaw/ucantone/executor"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/container"
	"github.com/alanshaw/ucantone/ucan/receipt"
)

type HTTPExecRequest struct {
	Request    *http.Request
	invocation ucan.Invocation
	metadata   ucan.Container
}

func (h *HTTPExecRequest) Context() context.Context {
	panic("unimplemented")
}

func (h *HTTPExecRequest) Invocation() ucan.Invocation {
	return h.invocation
}

func (h *HTTPExecRequest) Metadata() ucan.Container {
	return h.metadata
}

func (h *HTTPExecRequest) Task() ucan.Task {
	return h.invocation.Task()
}

var _ executor.ExecutionRequest = (*HTTPExecRequest)(nil)

type HTTPExecResponse struct {
	result   result.Result[ipld.Any, error]
	metadata ucan.Container
}

func (h *HTTPExecResponse) SetMetadata(meta ucan.Container) error {
	h.metadata = meta
	return nil
}

func (h *HTTPExecResponse) SetResult(o ipld.Any, x error) error {
	if x != nil {
		h.result = result.Error[ipld.Any, error](x)
	} else {
		h.result = result.OK[ipld.Any, error](o)
	}
	return nil
}

var _ executor.ExecutionResponse = (*HTTPExecResponse)(nil)

type Server struct {
	ID       ucan.Signer
	Executor executor.Executor
}

func New(executor executor.Executor) *Server {
	return &Server{
		Executor: executor,
	}
}

// RoundTrip unpacks and executes an incoming request, returning the response.
func (s *Server) RoundTrip(r *http.Request) (*http.Response, error) {
	buf, err := io.ReadAll(r.Body)
	ct, err := container.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("decoding container: %w", err)
	}

	var receipts []ucan.Receipt
	var invocations []ucan.Invocation
	var delegations []ucan.Delegation
	for _, inv := range ct.Invocations() {
		req := HTTPExecRequest{r, inv, ct}
		res := HTTPExecResponse{}
		err := s.Executor.Execute(&req, &res)
		if err != nil {
			// This shouldn't really happen, executor only returns an error when
			// result or  metadata cannot be set, which is likely a developer error.
			return nil, fmt.Errorf("executing task %s: %w", inv.Task().Link(), err)
		}
		receipt, err := receipt.Issue(
			s.ID,
			inv.Link(),
			out,
			receipt.WithCause(inv.Link()),
		)
		if err != nil {
			return nil, fmt.Errorf("issuing receipt for task %s: %w", inv.Task().Link(), err)
		}
		receipts = append(receipts, receipt)
	}
}
