package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/execution/executor"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/container"
	"github.com/alanshaw/ucantone/ucan/receipt"
)

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
		req := execution.NewRequest(r.Context(), inv, reqContainer)
		res := execution.NewResponse()

		err := s.Executor.Execute(req, res)
		if err != nil {
			// This shouldn't really happen, executor only returns an error when
			// result or metadata cannot be set, which is likely a developer error.
			return nil, fmt.Errorf("executing task %s: %w", inv.Task().Link(), err)
		}

		receipt, err := receipt.Issue(
			s.ID,
			inv.Link(),
			res.Result(),
			receipt.WithCause(inv.Link()),
		)
		if err != nil {
			return nil, fmt.Errorf("issuing receipt for task %s: %w", inv.Task().Link(), err)
		}
		receipts = append(receipts, receipt)

		if res.Metadata() != nil {
			invocations = append(invocations, res.Metadata().Invocations()...)
			delegations = append(delegations, res.Metadata().Delegations()...)
			receipts = append(receipts, res.Metadata().Receipts()...)
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
