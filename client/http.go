package client

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/transport"
)

type HTTPClient struct {
	*Client[*http.Request, *http.Response]
}

func NewHTTP(serviceURL *url.URL, options ...HTTPOption) (*HTTPClient, error) {
	cfg := httpClientConfig{
		codec:  transport.DefaultHTTPOutboundCodec,
		client: http.DefaultClient,
	}
	for _, opt := range options {
		opt(&cfg)
	}
	return &HTTPClient{
		Client: New(&httpTransport{cfg.client, serviceURL}, cfg.codec),
	}, nil
}

func (c *HTTPClient) Execute(execRequest execution.Request) (execution.Response, error) {
	res, err := c.Client.Execute(execRequest)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	httpMeta, ok := res.Metadata().(*transport.HTTPResponseContainer)
	if !ok {
		return nil, fmt.Errorf("expected HTTPResponseContainer, got %T", res.Metadata())
	}
	err = httpMeta.Response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("closing body: %w", err)
	}
	return res, nil
}

type httpTransport struct {
	client *http.Client
	url    *url.URL
}

func (t *httpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Method = http.MethodPost
	r.URL = t.url
	return t.client.Do(r)
}
