package server

import (
	"bytes"
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

func (r *HTTPExecRequest) Context() context.Context {
	return r.Request.Context()
}

func (r *HTTPExecRequest) Invocation() ucan.Invocation {
	return r.invocation
}

func (r *HTTPExecRequest) Metadata() ucan.Container {
	return r.metadata
}

func (r *HTTPExecRequest) Task() ucan.Task {
	return r.invocation.Task()
}

var _ executor.ExecutionRequest = (*HTTPExecRequest)(nil)

type HTTPExecResponse struct {
	result   result.Result[ipld.Any, error]
	metadata ucan.Container
}

func (r *HTTPExecResponse) SetMetadata(meta ucan.Container) error {
	r.metadata = meta
	return nil
}

func (r *HTTPExecResponse) SetResult(o ipld.Any, x error) error {
	if x != nil {
		r.result = result.Error[ipld.Any, error](x)
	} else {
		r.result = result.OK[ipld.Any, error](o)
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
	if r.Header.Get("Content-Type") != "application/vnd.ipld.dag-cbor" {
		return nil, fmt.Errorf("invalid content type %s, expected application/vnd.ipld.dag-cbor", r.Header.Get("Content-Type"))
	}

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("decoding container: %w", err)
	}
	reqContainer, err := container.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("decoding container: %w", err)
	}

	var receipts []ucan.Receipt
	var invocations []ucan.Invocation
	var delegations []ucan.Delegation
	for _, inv := range reqContainer.Invocations() {
		req := HTTPExecRequest{r, inv, reqContainer}
		res := HTTPExecResponse{}

		err := s.Executor.Execute(&req, &res)
		if err != nil {
			// This shouldn't really happen, executor only returns an error when
			// result or metadata cannot be set, which is likely a developer error.
			return nil, fmt.Errorf("executing task %s: %w", inv.Task().Link(), err)
		}

		receipt, err := receipt.Issue(
			s.ID,
			inv.Link(),
			res.result,
			receipt.WithCause(inv.Link()),
		)
		if err != nil {
			return nil, fmt.Errorf("issuing receipt for task %s: %w", inv.Task().Link(), err)
		}
		receipts = append(receipts, receipt)

		if res.metadata != nil {
			invocations = append(invocations, res.metadata.Invocations()...)
			delegations = append(delegations, res.metadata.Delegations()...)
			receipts = append(receipts, res.metadata.Receipts()...)
		}
	}

	respContainer, err := container.New(
		container.WithInvocations(invocations...),
		container.WithDelegations(delegations...),
		container.WithReceipts(receipts...),
	)
	if err != nil {
		return nil, fmt.Errorf("creating response container: %w", err)
	}

	respBuf, err := container.Encode(container.Raw, respContainer)
	if err != nil {
		return nil, fmt.Errorf("encoding response container: %w", err)
	}

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(respBuf)),
		Header:     make(http.Header),
	}
	resp.Header.Set("Content-Type", "application/vnd.ipld.dag-cbor")
	return resp, nil
}
