package client

import (
	"fmt"

	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/transport"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/container"
)

type Client[Req transport.Request, Res any] struct {
	Codec     transport.OutboundCodec[Req, Res]
	Transport transport.RoundTripper[Req, Res]
}

func New[Req transport.Request, Res any](transport transport.RoundTripper[Req, Res], codec transport.OutboundCodec[Req, Res]) *Client[Req, Res] {
	return &Client[Req, Res]{
		Transport: transport,
		Codec:     codec,
	}
}

func (c *Client[Req, Res]) Execute(execRequest execution.Request) (execution.Response, error) {
	invocations := []ucan.Invocation{execRequest.Invocation()}
	var delegations []ucan.Delegation
	var receipts []ucan.Receipt
	if execRequest.Metadata() != nil {
		invocations = append(invocations, execRequest.Metadata().Invocations()...)
		delegations = append(delegations, execRequest.Metadata().Delegations()...)
		receipts = append(receipts, execRequest.Metadata().Receipts()...)
	}
	reqContainer := container.New(
		container.WithInvocations(invocations...),
		container.WithDelegations(delegations...),
		container.WithReceipts(receipts...),
	)
	request, err := c.Codec.Encode(reqContainer)
	if err != nil {
		return nil, fmt.Errorf("encoding container: %w", err)
	}
	response, err := c.Transport.RoundTrip(request)
	if err != nil {
		return nil, fmt.Errorf("roundtripping request: %w", err)
	}
	resContainer, err := c.Codec.Decode(response)
	if err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	task := execRequest.Invocation().Task()
	var receipt ucan.Receipt
	// find receipt for the invocation task
	for _, r := range resContainer.Receipts() {
		if r.Ran() == task.Link() {
			receipt = r
			break
		}
	}
	if receipt == nil {
		return nil, fmt.Errorf("missing receipt for task: %s", task.Link())
	}
	return execution.NewResponse(
		task.Link(),
		execution.WithReceipt(receipt),
		execution.WithMetadata(resContainer),
	)
}
