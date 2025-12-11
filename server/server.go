package server

import (
	"fmt"
	"net/http"

	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/execution/dispatcher"
	"github.com/alanshaw/ucantone/principal"
	"github.com/alanshaw/ucantone/transport"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/container"
	"github.com/alanshaw/ucantone/ucan/receipt"
	"github.com/alanshaw/ucantone/validator"
)

type HTTPServer struct {
	id       principal.Signer
	executor *dispatcher.Dispatcher
	codec    transport.InboundCodec[*http.Request, *http.Response]
}

// NewHTTP creates a new UCAN HTTP server capable of handling UCAN invocations
// over HTTP.
func NewHTTP(id principal.Signer, options ...Option) *HTTPServer {
	cfg := serverConfig{
		codec: transport.DefaultHTTPInboundCodec,
	}
	for _, opt := range options {
		opt(&cfg)
	}
	executor := dispatcher.New(
		id.Verifier(),
		dispatcher.WithValidationOptions(cfg.validationOpts...),
	)
	return &HTTPServer{
		id:       id,
		codec:    cfg.codec,
		executor: executor,
	}
}

func (s *HTTPServer) Handle(capability validator.Capability, fn execution.HandlerFunc) {
	s.executor.Handle(capability, fn)
}

// RoundTrip unpacks and executes an incoming request, returning the response.
func (s *HTTPServer) RoundTrip(r *http.Request) (*http.Response, error) {
	reqContainer, err := s.codec.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("decoding request: %w", err)
	}

	var invocations []ucan.Invocation
	var delegations []ucan.Delegation
	var receipts []ucan.Receipt
	for _, inv := range reqContainer.Invocations() {
		req := execution.NewRequest(r.Context(), inv, reqContainer)

		res, err := s.executor.Execute(req)
		if err != nil {
			// This shouldn't really happen, executor only returns an error when
			// result or metadata cannot be set, which is likely a developer error.
			return nil, fmt.Errorf("executing task %s: %w", inv.Task().Link(), err)
		}

		receipt, err := receipt.Issue(
			s.id,
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

	resp, err := s.codec.Encode(respContainer)
	if err != nil {
		return nil, fmt.Errorf("encoding response container: %w", err)
	}

	return resp, nil
}
